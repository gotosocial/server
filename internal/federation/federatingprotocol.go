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
	"github.com/go-fed/activity/streams/vocab"
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/db/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/transport"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

// federatingProtocol implements the go-fed federating protocol interface
type federatingProtocol struct {
	db                  db.DB
	log                 *logrus.Logger
	config              *config.Config
	transportController transport.Controller
	typeConverter       typeutils.TypeConverter
}

// newFederatingProtocol returns the gotosocial implementation of the GTSFederatingProtocol interface
func newFederatingProtocol(db db.DB, log *logrus.Logger, config *config.Config, transportController transport.Controller, typeConverter typeutils.TypeConverter) pub.FederatingProtocol {
	return &federatingProtocol{
		db:                  db,
		log:                 log,
		config:              config,
		transportController: transportController,
		typeConverter: typeConverter,
	}
}

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
func (f *federatingProtocol) PostInboxRequestBodyHook(ctx context.Context, r *http.Request, activity pub.Activity) (context.Context, error) {
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

	ctxWithActivity := context.WithValue(ctx, util.APActivity, activity)
	return ctxWithActivity, nil
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
func (f *federatingProtocol) AuthenticatePostInbox(ctx context.Context, w http.ResponseWriter, r *http.Request) (context.Context, bool, error) {
	l := f.log.WithFields(logrus.Fields{
		"func":      "AuthenticatePostInbox",
		"useragent": r.UserAgent(),
		"url":       r.URL.String(),
	})
	l.Trace("received request to authenticate")

	requestedAccountI := ctx.Value(util.APAccount)
	if requestedAccountI == nil {
		return ctx, false, errors.New("requested account not set in context")
	}

	requestedAccount, ok := requestedAccountI.(*gtsmodel.Account)
	if !ok || requestedAccount == nil {
		return ctx, false, errors.New("requested account not parsebale from context")
	}

	transport, err := f.transportController.NewTransport(requestedAccount.PublicKeyURI, requestedAccount.PrivateKey)
	if err != nil {
		return ctx, false, fmt.Errorf("error creating transport: %s", err)
	}

	requestingPublicKeyID, err := AuthenticateFederatedRequest(transport, r)
	if err != nil {
		l.Debugf("request not authenticated: %s", err)
		return ctx, false, fmt.Errorf("not authenticated: %s", err)
	}

	requestingAccount := &gtsmodel.Account{}
	if err := f.db.GetWhere("public_key_uri", requestingPublicKeyID.String(), requestingAccount); err != nil {
		// there's been a proper error so return it
		if _, ok := err.(db.ErrNoEntries); !ok {
			return ctx, false, fmt.Errorf("error getting requesting account with public key id %s: %s", requestingPublicKeyID.String(), err)
		}
		// we just don't know this account (yet) so try to dereference it
		// TODO: slow-fed
		person, err := DereferenceAccount(transport, requestingPublicKeyID)
		if err != nil {
			return ctx, false, fmt.Errorf("error dereferencing account with public key id %s: %s", requestingPublicKeyID.String(), err)
		}
		a, err := f.typeConverter.ASPersonToAccount(person)
		if err != nil {
			return ctx, false, fmt.Errorf("error converting person with public key id %s to account: %s", requestingPublicKeyID.String(), err)
		}
		requestingAccount = a
	}

	return newContext, true, nil
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
func (f *federatingProtocol) Blocked(ctx context.Context, actorIRIs []*url.URL) (bool, error) {
	// TODO
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
func (f *federatingProtocol) FederatingCallbacks(ctx context.Context) (pub.FederatingWrappedCallbacks, []interface{}, error) {
	// TODO
	return pub.FederatingWrappedCallbacks{}, nil, nil
}

// DefaultCallback is called for types that go-fed can deserialize but
// are not handled by the application's callbacks returned in the
// Callbacks method.
//
// Applications are not expected to handle every single ActivityStreams
// type and extension, so the unhandled ones are passed to
// DefaultCallback.
func (f *federatingProtocol) DefaultCallback(ctx context.Context, activity pub.Activity) error {
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
func (f *federatingProtocol) MaxInboxForwardingRecursionDepth(ctx context.Context) int {
	// TODO
	return 0
}

// MaxDeliveryRecursionDepth determines how deep to search within
// collections owned by peers when they are targeted to receive a
// delivery.
//
// Zero or negative numbers indicate infinite recursion.
func (f *federatingProtocol) MaxDeliveryRecursionDepth(ctx context.Context) int {
	// TODO
	return 0
}

// FilterForwarding allows the implementation to apply business logic
// such as blocks, spam filtering, and so on to a list of potential
// Collections and OrderedCollections of recipients when inbox
// forwarding has been triggered.
//
// The activity is provided as a reference for more intelligent
// logic to be used, but the implementation must not modify it.
func (f *federatingProtocol) FilterForwarding(ctx context.Context, potentialRecipients []*url.URL, a pub.Activity) ([]*url.URL, error) {
	// TODO
	return nil, nil
}

// GetInbox returns the OrderedCollection inbox of the actor for this
// context. It is up to the implementation to provide the correct
// collection for the kind of authorization given in the request.
//
// AuthenticateGetInbox will be called prior to this.
//
// Always called, regardless whether the Federated Protocol or Social
// API is enabled.
func (f *federatingProtocol) GetInbox(ctx context.Context, r *http.Request) (vocab.ActivityStreamsOrderedCollectionPage, error) {
	// IMPLEMENTATION NOTE: For GoToSocial, we serve outboxes and inboxes through
	// the CLIENT API, not through the federation API, so we just do nothing here.
	return nil, nil
}