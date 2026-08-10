// Package item implements the resource exposed by the service: its model, its
// DynamoDB repository and the use cases on top of it.
package item

import "time"

// Item is the resource stored in DynamoDB and returned by the API.
type Item struct {
	ID        string            `json:"id" dynamodbav:"id"`
	Name      string            `json:"name" dynamodbav:"name"`
	Tags      []string          `json:"tags,omitempty" dynamodbav:"tags,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty" dynamodbav:"metadata,omitempty"`
	CreatedAt time.Time         `json:"createdAt" dynamodbav:"createdAt"`
}
