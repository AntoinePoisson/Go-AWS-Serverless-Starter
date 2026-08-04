"use strict";

const routes = [
  { method: "GET", path: "/health" },
  { method: "GET", path: "/public/items/{id}" },
];

// Nothing on top of provider.environment: this function holds no secret.
const environment = {};

const permissions = [
  {
    Effect: "Allow",
    Action: ["dynamodb:GetItem"],
    Resource: { "Fn::GetAtt": ["ItemsTable", "Arn"] },
  },
];

module.exports = {
  public: {
    handler: "bootstrap",
    description: "Public read-only endpoints",
    package: {
      artifact: "bin/dist/public.zip",
    },
    environment,
    iamRoleStatements: permissions,
    events: routes.map((route) => ({ httpApi: route })),
  },
};
