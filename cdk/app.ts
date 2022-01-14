import * as apigw from "@aws-cdk/aws-apigatewayv2";
import * as integrations from "@aws-cdk/aws-apigatewayv2-integrations";
import * as cdk from "@aws-cdk/core";
import { Duration } from "@aws-cdk/core";
import Function from "./constructs/function";
import * as secretsmanager from "@aws-cdk/aws-secretsmanager";
import * as dynamodb from "@aws-cdk/aws-dynamodb";

class Stack extends cdk.Stack {
  constructor(scope: cdk.App, id: string) {
    super(scope, id);

    // Create DynamoDb table
    const table = new dynamodb.Table(this, 'random-notion-cache', {
      partitionKey: { name: 'database_id', type: dynamodb.AttributeType.STRING },
    });

    // Create API and link to DynamoDb
    const logLevel = 1;  // Debug = 0, Info = 1, Trace = -1
    const apiHandler = new Function(this, "RandomNotionFunction", logLevel, { 
      entry: "../go/cmd/lambda",
      moduleDir: "../go/go.mod",
      timeout: Duration.seconds(30),
      environment: {
        CACHE_TABLE_NAME: table.tableName,
      },
    });
    const api = new apigw.HttpApi(this, "RandomNotionAPI");

    api.addRoutes({
      path: "/",
      methods: [apigw.HttpMethod.GET],
      integration: new integrations.LambdaProxyIntegration({
        handler: apiHandler,
      })
    });

    // Grant access to AWS Secret Manager
    const apiKeySecretArn = "arn:aws:secretsmanager:us-west-2:760655967349:secret:random-notion/notion-api-zFj6xG";
    const apiKeySecret = secretsmanager.Secret.fromSecretCompleteArn(
      this,
      'SecretFromCompleteArn',
      apiKeySecretArn
    );
    apiKeySecret.grantRead(apiHandler);

    // Grant access to DynamoDb table
    table.grantReadWriteData(apiHandler);
  }
}

const app = new cdk.App();
new Stack(app, "RandomNotion");