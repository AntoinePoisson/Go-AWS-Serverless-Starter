// Package item implements the resource this service exposes: model, DynamoDB
// repository and the use cases on top.
package item

import "time"

// Item is what we store in DynamoDB and return from the API.
type Item struct {
	ID        string            `json:"id" dynamodbav:"id" validate:"required" format:"uuid" example:"550e8400-e29b-41d4-a716-446655440000"`
	Name      string            `json:"name" dynamodbav:"name" validate:"required" maxLength:"200" example:"first item"`
	Tags      []string          `json:"tags,omitempty" dynamodbav:"tags,omitempty" validate:"max=20" maxLength:"50" example:"demo,starter"`
	Metadata  map[string]string `json:"metadata,omitempty" dynamodbav:"metadata,omitempty" example:"owner:platform"`
	CreatedAt time.Time         `json:"createdAt" dynamodbav:"createdAt" validate:"required" format:"date-time" example:"2026-08-27T10:00:00Z"`
}
