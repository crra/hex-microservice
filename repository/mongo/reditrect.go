package mongo

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type redirect struct {
	ID        primitive.ObjectID `json:"ID" bson:"_id,omitempty"`
	Code      string             `json:"code" bson:"code"`
	URL       string             `json:"url" bson:"url"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
}
