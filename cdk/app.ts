import * as apigw from "@aws-cdk/aws-apigatewayv2";
import * as integrations from "@aws-cdk/aws-apigatewayv2-integrations";
import * as cdk from "@aws-cdk/core";
import { Duration } from "@aws-cdk/core";
import Function from "./constructs/function";

class Stack extends cdk.Stack {
  constructor(scope: cdk.App, id: string) {
    super(scope, id);

    const apiHandler = new Function(this, "RandomNotionFunction", { 
      entry: "..",
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
  }
}

const app = new cdk.App();
new Stack(app, "RandomNotion");