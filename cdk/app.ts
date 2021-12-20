import * as apigw from "@aws-cdk/aws-apigatewayv2";
import * as integrations from "@aws-cdk/aws-apigatewayv2-integrations";
import * as cdk from "@aws-cdk/core";
import { Duration } from "@aws-cdk/core";
import Function from "./constructs/function";
import * as secretsmanager from '@aws-cdk/aws-secretsmanager';

class Stack extends cdk.Stack {
  constructor(scope: cdk.App, id: string) {
    super(scope, id);

    const apiHandler = new Function(this, "RandomNotionFunction", { 
      entry: "../go/cmd/lambda",
      moduleDir: "../go/go.mod",
      timeout: Duration.seconds(30),
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
  }
}

const app = new cdk.App();
new Stack(app, "RandomNotion");