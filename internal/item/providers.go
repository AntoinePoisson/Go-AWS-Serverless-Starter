package item

import "github.com/google/wire"

// WireSet wires the DynamoDB-backed item service. Binding DynamoDBAPI to a
// real client is up to each function.
var WireSet = wire.NewSet(
	NewDynamoDBRepository,
	wire.Bind(new(Repository), new(*DynamoDBRepository)),
	NewService,
	wire.Bind(new(ServiceAPI), new(*Service)),
)
