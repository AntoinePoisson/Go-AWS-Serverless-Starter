package item

import "github.com/google/wire"

// WireSet is the DynamoDB-backed item service. Each function binds DynamoDBAPI
// to a real client.
var WireSet = wire.NewSet(
	NewDynamoDBRepository,
	wire.Bind(new(Repository), new(*DynamoDBRepository)),
	NewService,
	wire.Bind(new(ServiceAPI), new(*Service)),
)
