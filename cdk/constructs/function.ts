import * as lambda from "@aws-cdk/aws-lambda";
import * as golambda from "@aws-cdk/aws-lambda-go";
import * as cdk from "@aws-cdk/core";

export default class Function extends golambda.GoFunction {
  constructor(
    scope: cdk.Construct,
    id: string,
    props: golambda.GoFunctionProps
  ) {
    props = {
      tracing: lambda.Tracing.ACTIVE,
      insightsVersion: lambda.LambdaInsightsVersion.VERSION_1_0_98_0,
      ...props,
    };
    super(scope, id, props);
  }
}
