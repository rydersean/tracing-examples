package database

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo/bson"
	"github.com/opentracing/opentracing-go"
	mgotrace "github.com/signalfx/signalfx-go-tracing/contrib/globalsign/mgo"

	"github.com/signalfx/tracing-examples/signalfx-tracing/signalfx-go-tracing/gin/server/models"
)

type mgoManager struct {
	Host        string
	Port        int
	Name        string
	ServiceName string
}

var _ Manager = (*mgoManager)(nil)

var mgoSession *mgotrace.Session

// GetBoardByID returns a board for a given boardID
func (m *mgoManager) GetBoardByID(c *gin.Context, id string) (models.Board, error) {
	collection := m.getCollection(c)

	board := models.Board{}
	err := collection.Find(bson.M{"board_id": id}).One(&board)
	if err != nil {
		return models.Board{}, err
	}
	return board, nil
}

// InsertBoard inserts a given board
func (m *mgoManager) InsertBoard(c *gin.Context, board models.Board) error {
	collection := m.getCollection(c)

	return collection.Insert(board)
}

// UpdateBoard saves a given updated board
func (m *mgoManager) UpdateBoard(c *gin.Context, board models.Board) error {
	collection := m.getCollection(c)

	return collection.Update(bson.M{"board_id": board.ID}, board)
}

// getCollection returns board collection
func (m *mgoManager) getCollection(c *gin.Context) *mgotrace.Collection {
	if mgoSession == nil {
		parentSpan, ctx := opentracing.StartSpanFromContext(c.Request.Context(), "mongo.session")
		c.Set("parentSpan", parentSpan)

		var err error
		mgoSession, err = mgotrace.Dial(fmt.Sprintf("mongodb://%s:%d/%s", m.Host, m.Port, m.Name), mgotrace.WithServiceName(m.ServiceName), mgotrace.WithContext(ctx))

		if err != nil {
			fmt.Printf("Can't connect to mongo, go error %v\n", err)
			panic(err.Error())
		}
	}

	db := mgoSession.DB(m.Name)
	collection := db.C(models.CollectionBoard)

	return collection
}

// Close closes open DB session
func (m *mgoManager) Close(c *gin.Context) {
	if mgoSession != nil {
		mgoSession.Close()
		mgoSession = nil
		if value, exists := c.Get("parentSpan"); exists {
			if parentSpan, ok := value.(opentracing.Span); ok {
				parentSpan.Finish()
			}
		}
	}
}
