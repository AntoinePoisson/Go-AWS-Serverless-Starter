// Package item implements the resource this service exposes: model, DynamoDB
// repository and the use cases on top.
package item

import "time"

// Item is what we store in DynamoDB and return from the API.
type Item struct {
	ID        string            `json:"id" dynamodbav:"id"`
	Name      string            `json:"name" dynamodbav:"name"`
	Tags      []string          `json:"tags,omitempty" dynamodbav:"tags,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty" dynamodbav:"metadata,omitempty"`
	CreatedAt time.Time         `json:"createdAt" dynamodbav:"createdAt"`
}
