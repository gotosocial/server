package streaming

import (
	"sync"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
	"github.com/superseriousbusiness/gotosocial/internal/visibility"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
)

// Processor wraps a bunch of functions for processing streaming.
type Processor interface {
	// AuthorizeStreamingRequest returns an oauth2 token info in response to an access token query from the streaming API
	AuthorizeStreamingRequest(accessToken string) (*gtsmodel.Account, error)
	OpenStreamForAccount(c *websocket.Conn, account *gtsmodel.Account, streamType string) gtserror.WithCode
	StreamStatusForAccount(s *apimodel.Status, account *gtsmodel.Account) error
}

type processor struct {
	tc          typeutils.TypeConverter
	config      *config.Config
	db          db.DB
	filter      visibility.Filter
	log         *logrus.Logger
	oauthServer oauth.Server
	streamMap   *sync.Map
}

// New returns a new status processor.
func New(db db.DB, tc typeutils.TypeConverter, oauthServer oauth.Server, config *config.Config, log *logrus.Logger) Processor {
	return &processor{
		tc:          tc,
		config:      config,
		db:          db,
		filter:      visibility.NewFilter(db, log),
		log:         log,
		oauthServer: oauthServer,
		streamMap:   &sync.Map{},
	}
}
