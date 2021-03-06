/*
   GoToSocial
   Copyright (C) 2021 GoToSocial Authors admin@gotosocial.org

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package federation

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/go-fed/activity/pub"
	"github.com/go-fed/activity/streams"
	"github.com/go-fed/activity/streams/vocab"
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

/*
	GO FED FEDERATING PROTOCOL INTERFACE
	FederatingProtocol contains behaviors an application needs to satisfy for the
	full ActivityPub S2S implementation to be supported by this library.
	It is only required if the client application wants to support the server-to-
	server, or federating, protocol.
	It is passed to the library as a dependency injection from the client
	application.
*/

// PostInboxRequestBodyHook callback after parsing the request body for a federated request
// to the Actor's inbox.
//
// Can be used to set contextual information based on the Activity
// received.
//
// Only called if the Federated Protocol is enabled.
//
// Warning: Neither authentication nor authorization has taken place at
// this time. Doing anything beyond setting contextual information is
// strongly discouraged.
//
// If an error is returned, it is passed back to the caller of
// PostInbox. In this case, the DelegateActor implementation must not
// write a response to the ResponseWriter as is expected that the caller
// to PostInbox will do so when handling the error.
func (f *federator) PostInboxRequestBodyHook(ctx context.Context, r *http.Request, activity pub.Activity) (context.Context, error) {
	l := f.log.WithFields(logrus.Fields{
		"func":      "PostInboxRequestBodyHook",
		"useragent": r.UserAgent(),
		"url":       r.URL.String(),
	})

	if activity == nil {
		err := errors.New("nil activity in PostInboxRequestBodyHook")
		l.Debug(err)
		return nil, err
	}
	// set the activity on the context for use later on
	return context.WithValue(ctx, util.APActivity, activity), nil
}

// AuthenticatePostInbox delegates the authentication of a POST to an
// inbox.
//
// If an error is returned, it is passed back to the caller of
// PostInbox. In this case, the implementation must not write a
// response to the ResponseWriter as is expected that the client will
// do so when handling the error. The 'authenticated' is ignored.
//
// If no error is returned, but authentication or authorization fails,
// then authenticated must be false and error nil. It is expected that
// the implementation handles writing to the ResponseWriter in this
// case.
//
// Finally, if the authentication and authorization succeeds, then
// authenticated must be true and error nil. The request will continue
// to be processed.
func (f *federator) AuthenticatePostInbox(ctx context.Context, w http.ResponseWriter, r *http.Request) (context.Context, bool, error) {
	l := f.log.WithFields(logrus.Fields{
		"func":      "AuthenticatePostInbox",
		"useragent": r.UserAgent(),
		"url":       r.URL.String(),
	})
	l.Trace("received request to authenticate")

	if !util.IsInboxPath(r.URL) {
		return nil, false, fmt.Errorf("path %s was not an inbox path", r.URL.String())
	}

	username, err := util.ParseInboxPath(r.URL)
	if err != nil {
		return nil, false, fmt.Errorf("could not parse path %s: %s", r.URL.String(), err)
	}

	if username == "" {
		return nil, false, errors.New("username was empty")
	}

	requestedAccount := &gtsmodel.Account{}
	if err := f.db.GetLocalAccountByUsername(username, requestedAccount); err != nil {
		return nil, false, fmt.Errorf("could not fetch requested account with username %s: %s", username, err)
	}

	publicKeyOwnerURI, authenticated, err := f.AuthenticateFederatedRequest(ctx, requestedAccount.Username)
	if err != nil {
		l.Debugf("request not authenticated: %s", err)
		return ctx, false, err
	}

	if !authenticated {
		w.WriteHeader(http.StatusForbidden)
		return ctx, false, nil
	}

	// authentication has passed, so add an instance entry for this instance if it hasn't been done already
	i := &gtsmodel.Instance{}
	if err := f.db.GetWhere([]db.Where{{Key: "domain", Value: publicKeyOwnerURI.Host, CaseInsensitive: true}}, i); err != nil {
		if _, ok := err.(db.ErrNoEntries); !ok {
			// there's been an actual error
			return ctx, false, fmt.Errorf("error getting requesting account with public key id %s: %s", publicKeyOwnerURI.String(), err)
		}

		// we don't have an entry for this instance yet so dereference it
		i, err = f.DereferenceRemoteInstance(username, &url.URL{
			Scheme: publicKeyOwnerURI.Scheme,
			Host:   publicKeyOwnerURI.Host,
		})
		if err != nil {
			return nil, false, fmt.Errorf("could not dereference new remote instance %s during AuthenticatePostInbox: %s", publicKeyOwnerURI.Host, err)
		}

		// and put it in the db
		if err := f.db.Put(i); err != nil {
			return nil, false, fmt.Errorf("error inserting newly dereferenced instance %s: %s", publicKeyOwnerURI.Host, err)
		}
	}

	requestingAccount := &gtsmodel.Account{}
	if err := f.db.GetWhere([]db.Where{{Key: "uri", Value: publicKeyOwnerURI.String()}}, requestingAccount); err != nil {
		// there's been a proper error so return it
		if _, ok := err.(db.ErrNoEntries); !ok {
			return ctx, false, fmt.Errorf("error getting requesting account with public key id %s: %s", publicKeyOwnerURI.String(), err)
		}

		// we don't know this account (yet) so let's dereference it right now
		person, err := f.DereferenceRemoteAccount(requestedAccount.Username, publicKeyOwnerURI)
		if err != nil {
			return ctx, false, fmt.Errorf("error dereferencing account with public key id %s: %s", publicKeyOwnerURI.String(), err)
		}

		a, err := f.typeConverter.ASRepresentationToAccount(person, false)
		if err != nil {
			return ctx, false, fmt.Errorf("error converting person with public key id %s to account: %s", publicKeyOwnerURI.String(), err)
		}

		aID, err := id.NewRandomULID()
		if err != nil {
			return ctx, false, err
		}
		a.ID = aID

		if err := f.db.Put(a); err != nil {
			l.Errorf("error inserting dereferenced remote account: %s", err)
		}

		requestingAccount = a

		// send the newly dereferenced account into the processor channel for further async processing
		fromFederatorChanI := ctx.Value(util.APFromFederatorChanKey)
		if fromFederatorChanI == nil {
			l.Error("from federator channel wasn't set on context")
		}
		fromFederatorChan, ok := fromFederatorChanI.(chan gtsmodel.FromFederator)
		if !ok {
			l.Error("from federator channel was set on context but couldn't be parsed")
		}

		fromFederatorChan <- gtsmodel.FromFederator{
			APObjectType:   gtsmodel.ActivityStreamsProfile,
			APActivityType: gtsmodel.ActivityStreamsCreate,
			GTSModel:       requestingAccount,
		}
	}

	withRequester := context.WithValue(ctx, util.APRequestingAccount, requestingAccount)
	withRequested := context.WithValue(withRequester, util.APAccount, requestedAccount)
	return withRequested, true, nil
}

// Blocked should determine whether to permit a set of actors given by
// their ids are able to interact with this particular end user due to
// being blocked or other application-specific logic.
//
// If an error is returned, it is passed back to the caller of
// PostInbox.
//
// If no error is returned, but authentication or authorization fails,
// then blocked must be true and error nil. An http.StatusForbidden
// will be written in the wresponse.
//
// Finally, if the authentication and authorization succeeds, then
// blocked must be false and error nil. The request will continue
// to be processed.
//
// TODO: implement domain block checking here as well
func (f *federator) Blocked(ctx context.Context, actorIRIs []*url.URL) (bool, error) {
	l := f.log.WithFields(logrus.Fields{
		"func": "Blocked",
	})
	l.Debugf("entering BLOCKED function with IRI list: %+v", actorIRIs)

	requestedAccountI := ctx.Value(util.APAccount)
	requestedAccount, ok := requestedAccountI.(*gtsmodel.Account)
	if !ok {
		f.log.Errorf("requested account not set on request context")
		return false, errors.New("requested account not set on request context, so couldn't determine blocks")
	}

	for _, uri := range actorIRIs {
		blockedDomain, err := f.blockedDomain(uri.Host)
		if err != nil {
			return false, fmt.Errorf("error checking domain block: %s", err)
		}
		if blockedDomain {
			return true, nil
		}

		requestingAccount := &gtsmodel.Account{}
		if err := f.db.GetWhere([]db.Where{{Key: "uri", Value: uri.String()}}, requestingAccount); err != nil {
			_, ok := err.(db.ErrNoEntries)
			if ok {
				// we don't have an entry for this account so it's not blocked
				// TODO: allow a different default to be set for this behavior
				continue
			}
			return false, fmt.Errorf("error getting account with uri %s: %s", uri.String(), err)
		}

		// check if requested account blocks requesting account
		if err := f.db.GetWhere([]db.Where{
			{Key: "account_id", Value: requestedAccount.ID},
			{Key: "target_account_id", Value: requestingAccount.ID},
		}, &gtsmodel.Block{}); err == nil {
			// a block exists
			return true, nil
		}
	}
	return false, nil
}

// FederatingCallbacks returns the application logic that handles
// ActivityStreams received from federating peers.
//
// Note that certain types of callbacks will be 'wrapped' with default
// behaviors supported natively by the library. Other callbacks
// compatible with streams.TypeResolver can be specified by 'other'.
//
// For example, setting the 'Create' field in the
// FederatingWrappedCallbacks lets an application dependency inject
// additional behaviors they want to take place, including the default
// behavior supplied by this library. This is guaranteed to be compliant
// with the ActivityPub Social protocol.
//
// To override the default behavior, instead supply the function in
// 'other', which does not guarantee the application will be compliant
// with the ActivityPub Social Protocol.
//
// Applications are not expected to handle every single ActivityStreams
// type and extension. The unhandled ones are passed to DefaultCallback.
func (f *federator) FederatingCallbacks(ctx context.Context) (wrapped pub.FederatingWrappedCallbacks, other []interface{}, err error) {
	wrapped = pub.FederatingWrappedCallbacks{
		// OnFollow determines what action to take for this particular callback
		// if a Follow Activity is handled.
		//
		// For our implementation, we always want to do nothing because we have internal logic for handling follows.
		OnFollow: pub.OnFollowDoNothing,
	}

	other = []interface{}{
		// override default undo behavior and trigger our own side effects
		func(ctx context.Context, undo vocab.ActivityStreamsUndo) error {
			return f.FederatingDB().Undo(ctx, undo)
		},
		// override default accept behavior and trigger our own side effects
		func(ctx context.Context, accept vocab.ActivityStreamsAccept) error {
			return f.FederatingDB().Accept(ctx, accept)
		},
		// override default announce behavior and trigger our own side effects
		func(ctx context.Context, announce vocab.ActivityStreamsAnnounce) error {
			return f.FederatingDB().Announce(ctx, announce)
		},
	}

	return
}

// DefaultCallback is called for types that go-fed can deserialize but
// are not handled by the application's callbacks returned in the
// Callbacks method.
//
// Applications are not expected to handle every single ActivityStreams
// type and extension, so the unhandled ones are passed to
// DefaultCallback.
func (f *federator) DefaultCallback(ctx context.Context, activity pub.Activity) error {
	l := f.log.WithFields(logrus.Fields{
		"func":   "DefaultCallback",
		"aptype": activity.GetTypeName(),
	})
	l.Debugf("received unhandle-able activity type so ignoring it")
	return nil
}

// MaxInboxForwardingRecursionDepth determines how deep to search within
// an activity to determine if inbox forwarding needs to occur.
//
// Zero or negative numbers indicate infinite recursion.
func (f *federator) MaxInboxForwardingRecursionDepth(ctx context.Context) int {
	// TODO
	return 4
}

// MaxDeliveryRecursionDepth determines how deep to search within
// collections owned by peers when they are targeted to receive a
// delivery.
//
// Zero or negative numbers indicate infinite recursion.
func (f *federator) MaxDeliveryRecursionDepth(ctx context.Context) int {
	// TODO
	return 4
}

// FilterForwarding allows the implementation to apply business logic
// such as blocks, spam filtering, and so on to a list of potential
// Collections and OrderedCollections of recipients when inbox
// forwarding has been triggered.
//
// The activity is provided as a reference for more intelligent
// logic to be used, but the implementation must not modify it.
func (f *federator) FilterForwarding(ctx context.Context, potentialRecipients []*url.URL, a pub.Activity) ([]*url.URL, error) {
	// TODO
	return []*url.URL{}, nil
}

// GetInbox returns the OrderedCollection inbox of the actor for this
// context. It is up to the implementation to provide the correct
// collection for the kind of authorization given in the request.
//
// AuthenticateGetInbox will be called prior to this.
//
// Always called, regardless whether the Federated Protocol or Social
// API is enabled.
func (f *federator) GetInbox(ctx context.Context, r *http.Request) (vocab.ActivityStreamsOrderedCollectionPage, error) {
	// IMPLEMENTATION NOTE: For GoToSocial, we serve GETS to outboxes and inboxes through
	// the CLIENT API, not through the federation API, so we just do nothing here.
	return streams.NewActivityStreamsOrderedCollectionPage(), nil
}
