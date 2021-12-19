import * as lambda from "@aws-cdk/aws-lambda";
import * as golambda from "@aws-cdk/aws-lambda-go";
import * as cdk from "@aws-cdk/core";
import * as child_process from 'child_process';

function readGitRevision(): string {
  return child_process.execSync('git rev-parse --short HEAD').toString();
}

export default class Function extends golambda.GoFunction {
  constructor(
    scope: cdk.Construct,
    id: string,
    props: golambda.GoFunctionProps
  ) {
    const rev = readGitRevision();
    const flags = [`-ldflags "-X main.CommitID=${rev}"`];
    props = {
      tracing: lambda.Tracing.ACTIVE,
      insightsVersion: lambda.LambdaInsightsVersion.VERSION_1_0_98_0,
      logRetention: 14,
      bundling: {
        goBuildFlags: flags
      },
      ...props,
    };
    super(scope, id, props);
  }
}
