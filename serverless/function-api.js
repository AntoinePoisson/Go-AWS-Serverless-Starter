"use strict";

const routes = [
  { method: "POST", path: "/items" },
  { method: "GET", path: "/items" },
  { method: "GET", path: "/items/{id}" },
  { method: "DELETE", path: "/items/{id}" },
];

// merged into provider.environment (what both functions share)
const environment = {
  API_KEY: "${self:custom.stage.apiKey}",
};

const permissions = [
  {
    Effect: "Allow",
    Action: [
      "dynamodb:PutItem",
      "dynamodb:GetItem",
      "dynamodb:DeleteItem",
      "dynamodb:Scan",
    ],
    Resource: { "Fn::GetAtt": ["ItemsTable", "Arn"] },
  },
];

module.exports = {
  api: {
    handler: "bootstrap",
    description: "Authenticated CRUD API",
    package: {
      artifact: "bin/dist/api.zip",
    },
    environment,
    iamRoleStatements: permissions,
    events: routes.map((route) => ({ httpApi: route })),
  },
};
