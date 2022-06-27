/*
Copyright 2022 Gravitational, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package aws

import (
	"bytes"
	"net/http"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws/credentials"
	v4 "github.com/aws/aws-sdk-go/aws/signer/v4"
	"github.com/stretchr/testify/require"
)

// signingNameToHostname is a map of AWS services' signing names to their
// hostnames.
var signingNameToHostname = map[string]string{
	"a4b":                                   "a4b.us-east-1.amazonaws.com",
	"access-analyzer":                       "access-analyzer.us-east-1.amazonaws.com",
	"account":                               "account.us-east-1.amazonaws.com",
	"acm":                                   "acm.us-east-1.amazonaws.com",
	"acm-pca":                               "acm-pca.us-east-1.amazonaws.com",
	"airflow":                               "airflow.us-east-1.amazonaws.com",
	"amplify":                               "amplify.us-east-1.amazonaws.com",
	"amplifybackend":                        "amplifybackend.us-east-1.amazonaws.com",
	"amplifyuibuilder":                      "amplifyuibuilder.us-east-1.amazonaws.com",
	"apigateway":                            "apigateway.us-east-1.amazonaws.com",
	"app-integrations":                      "app-integrations.us-east-1.amazonaws.com",
	"appconfig":                             "appconfig.us-east-1.amazonaws.com",
	"appconfigdata":                         "appconfigdata.us-east-1.amazonaws.com",
	"appflow":                               "appflow.us-east-1.amazonaws.com",
	"application-autoscaling":               "application-autoscaling.us-east-1.amazonaws.com",
	"application-cost-profiler":             "application-cost-profiler.us-east-1.amazonaws.com",
	"applicationinsights":                   "applicationinsights.us-east-1.amazonaws.com",
	"appmesh":                               "appmesh.us-east-1.amazonaws.com",
	"apprunner":                             "apprunner.us-east-1.amazonaws.com",
	"appstream":                             "appstream2.us-east-1.amazonaws.com",
	"appsync":                               "appsync.us-east-1.amazonaws.com",
	"aps":                                   "aps.us-east-1.amazonaws.com",
	"athena":                                "athena.us-east-1.amazonaws.com",
	"auditmanager":                          "auditmanager.us-east-1.amazonaws.com",
	"autoscaling":                           "autoscaling.us-east-1.amazonaws.com",
	"autoscaling-plans":                     "autoscaling-plans.us-east-1.amazonaws.com",
	"aws-marketplace":                       "catalog.marketplace.us-east-1.amazonaws.com",
	"awsiottwinmaker":                       "iottwinmaker.us-east-1.amazonaws.com",
	"awsmigrationhubstrategyrecommendation": "migrationhub-strategy.us-east-1.amazonaws.com",
	"awsmobilehubservice":                   "mobile.us-east-1.amazonaws.com",
	"awsproton20200720":                     "proton.us-east-1.amazonaws.com",
	"awsssooidc":                            "oidc.us-east-1.amazonaws.com",
	"awsssoportal":                          "portal.sso.us-east-1.amazonaws.com",
	"backup":                                "backup.us-east-1.amazonaws.com",
	"backup-gateway":                        "backup-gateway.us-east-1.amazonaws.com",
	"batch":                                 "batch.us-east-1.amazonaws.com",
	"braket":                                "braket.us-east-1.amazonaws.com",
	"budgets":                               "budgets.amazonaws.com",
	"ce":                                    "ce.us-east-1.amazonaws.com",
	"chime":                                 "chime.us-east-1.amazonaws.com",
	"cloud9":                                "cloud9.us-east-1.amazonaws.com",
	"cloudcontrolapi":                       "cloudcontrolapi.us-east-1.amazonaws.com",
	"clouddirectory":                        "clouddirectory.us-east-1.amazonaws.com",
	"cloudformation":                        "cloudformation.us-east-1.amazonaws.com",
	"cloudfront":                            "cloudfront.amazonaws.com",
	"cloudhsm":                              "cloudhsm.us-east-1.amazonaws.com",
	"cloudsearch":                           "cloudsearch.us-east-1.amazonaws.com",
	"cloudtrail":                            "cloudtrail.us-east-1.amazonaws.com",
	"codeartifact":                          "codeartifact.us-east-1.amazonaws.com",
	"codebuild":                             "codebuild.us-east-1.amazonaws.com",
	"codecommit":                            "codecommit.us-east-1.amazonaws.com",
	"codedeploy":                            "codedeploy.us-east-1.amazonaws.com",
	"codeguru-profiler":                     "codeguru-profiler.us-east-1.amazonaws.com",
	"codeguru-reviewer":                     "codeguru-reviewer.us-east-1.amazonaws.com",
	"codepipeline":                          "codepipeline.us-east-1.amazonaws.com",
	"codestar":                              "codestar.us-east-1.amazonaws.com",
	"codestar-connections":                  "codestar-connections.us-east-1.amazonaws.com",
	"codestar-notifications":                "codestar-notifications.us-east-1.amazonaws.com",
	"cognito-identity":                      "cognito-identity.us-east-1.amazonaws.com",
	"cognito-idp":                           "cognito-idp.us-east-1.amazonaws.com",
	"cognito-sync":                          "cognito-sync.us-east-1.amazonaws.com",
	"comprehend":                            "comprehend.us-east-1.amazonaws.com",
	"comprehendmedical":                     "comprehendmedical.us-east-1.amazonaws.com",
	"compute-optimizer":                     "compute-optimizer.us-east-1.amazonaws.com",
	"config":                                "config.us-east-1.amazonaws.com",
	"connect":                               "connect.us-east-1.amazonaws.com",
	"cur":                                   "cur.us-east-1.amazonaws.com",
	"databrew":                              "databrew.us-east-1.amazonaws.com",
	"dataexchange":                          "dataexchange.us-east-1.amazonaws.com",
	"datapipeline":                          "datapipeline.us-east-1.amazonaws.com",
	"datasync":                              "datasync.us-east-1.amazonaws.com",
	"dax":                                   "dax.us-east-1.amazonaws.com",
	"detective":                             "api.detective.us-east-1.amazonaws.com",
	"devicefarm":                            "devicefarm.us-east-1.amazonaws.com",
	"devops-guru":                           "devops-guru.us-east-1.amazonaws.com",
	"directconnect":                         "directconnect.us-east-1.amazonaws.com",
	"discovery":                             "discovery.us-east-1.amazonaws.com",
	"dlm":                                   "dlm.us-east-1.amazonaws.com",
	"dms":                                   "dms.us-east-1.amazonaws.com",
	"drs":                                   "drs.us-east-1.amazonaws.com",
	"ds":                                    "ds.us-east-1.amazonaws.com",
	"dynamodb":                              "dynamodb.us-east-1.amazonaws.com",
	"ebs":                                   "ebs.us-east-1.amazonaws.com",
	"ec2":                                   "ec2.us-east-1.amazonaws.com",
	"ec2-instance-connect":                  "ec2-instance-connect.us-east-1.amazonaws.com",
	"ecr":                                   "api.ecr.us-east-1.amazonaws.com",
	"ecr-public":                            "api.ecr-public.us-east-1.amazonaws.com",
	"ecs":                                   "ecs.us-east-1.amazonaws.com",
	"eks":                                   "eks.us-east-1.amazonaws.com",
	"elastic-inference":                     "api.elastic-inference.us-east-1.amazonaws.com",
	"elasticache":                           "elasticache.us-east-1.amazonaws.com",
	"elasticbeanstalk":                      "elasticbeanstalk.us-east-1.amazonaws.com",
	"elasticfilesystem":                     "elasticfilesystem.us-east-1.amazonaws.com",
	"elasticloadbalancing":                  "elasticloadbalancing.us-east-1.amazonaws.com",
	"elasticmapreduce":                      "elasticmapreduce.us-east-1.amazonaws.com",
	"elastictranscoder":                     "elastictranscoder.us-east-1.amazonaws.com",
	"emr-containers":                        "emr-containers.us-east-1.amazonaws.com",
	"es":                                    "es.us-east-1.amazonaws.com",
	"events":                                "events.us-east-1.amazonaws.com",
	"evidently":                             "evidently.us-east-1.amazonaws.com",
	"finspace":                              "finspace.us-east-1.amazonaws.com",
	"finspace-api":                          "finspace-api.us-east-1.amazonaws.com",
	"firehose":                              "firehose.us-east-1.amazonaws.com",
	"fis":                                   "fis.us-east-1.amazonaws.com",
	"fms":                                   "fms.us-east-1.amazonaws.com",
	"forecast":                              "forecast.us-east-1.amazonaws.com",
	"frauddetector":                         "frauddetector.us-east-1.amazonaws.com",
	"fsx":                                   "fsx.us-east-1.amazonaws.com",
	"gamelift":                              "gamelift.us-east-1.amazonaws.com",
	"geo":                                   "geo.us-east-1.amazonaws.com",
	"glacier":                               "glacier.us-east-1.amazonaws.com",
	"globalaccelerator":                     "globalaccelerator.us-east-1.amazonaws.com",
	"glue":                                  "glue.us-east-1.amazonaws.com",
	"grafana":                               "grafana.us-east-1.amazonaws.com",
	"greengrass":                            "greengrass.us-east-1.amazonaws.com",
	"groundstation":                         "groundstation.us-east-1.amazonaws.com",
	"guardduty":                             "guardduty.us-east-1.amazonaws.com",
	"health":                                "health.us-east-1.amazonaws.com",
	"healthlake":                            "healthlake.us-east-1.amazonaws.com",
	"honeycode":                             "honeycode.us-east-1.amazonaws.com",
	"iam":                                   "iam.amazonaws.com",
	"identitystore":                         "identitystore.us-east-1.amazonaws.com",
	"imagebuilder":                          "imagebuilder.us-east-1.amazonaws.com",
	"inspector":                             "inspector.us-east-1.amazonaws.com",
	"inspector2":                            "inspector2.us-east-1.amazonaws.com",
	"iot":                                   "iot.us-east-1.amazonaws.com",
	"iot-jobs-data":                         "data.jobs.iot.us-east-1.amazonaws.com",
	"iot1click":                             "devices.iot1click.us-east-1.amazonaws.com",
	"iotanalytics":                          "iotanalytics.us-east-1.amazonaws.com",
	"iotdata":                               "data.iot.us-east-1.amazonaws.com",
	"iotdeviceadvisor":                      "api.iotdeviceadvisor.us-east-1.amazonaws.com",
	"iotevents":                             "iotevents.us-east-1.amazonaws.com",
	"ioteventsdata":                         "data.iotevents.us-east-1.amazonaws.com",
	"iotfleethub":                           "api.fleethub.iot.us-east-1.amazonaws.com",
	"iotsecuredtunneling":                   "api.tunneling.iot.us-east-1.amazonaws.com",
	"iotsitewise":                           "iotsitewise.us-east-1.amazonaws.com",
	"iotthingsgraph":                        "iotthingsgraph.us-east-1.amazonaws.com",
	"iotwireless":                           "api.iotwireless.us-east-1.amazonaws.com",
	"ivs":                                   "ivs.us-east-1.amazonaws.com",
	"kafka":                                 "kafka.us-east-1.amazonaws.com",
	"kafkaconnect":                          "kafkaconnect.us-east-1.amazonaws.com",
	"kendra":                                "kendra.us-east-1.amazonaws.com",
	"kinesis":                               "kinesis.us-east-1.amazonaws.com",
	"kinesisanalytics":                      "kinesisanalytics.us-east-1.amazonaws.com",
	"kinesisvideo":                          "kinesisvideo.us-east-1.amazonaws.com",
	"kms":                                   "kms.us-east-1.amazonaws.com",
	"lakeformation":                         "lakeformation.us-east-1.amazonaws.com",
	"lambda":                                "lambda.us-east-1.amazonaws.com",
	"lex":                                   "models-v2-lex.us-east-1.amazonaws.com",
	"license-manager":                       "license-manager.us-east-1.amazonaws.com",
	"lightsail":                             "lightsail.us-east-1.amazonaws.com",
	"logs":                                  "logs.us-east-1.amazonaws.com",
	"lookoutequipment":                      "lookoutequipment.us-east-1.amazonaws.com",
	"lookoutmetrics":                        "lookoutmetrics.us-east-1.amazonaws.com",
	"lookoutvision":                         "lookoutvision.us-east-1.amazonaws.com",
	"machinelearning":                       "machinelearning.us-east-1.amazonaws.com",
	"macie":                                 "macie.us-east-1.amazonaws.com",
	"macie2":                                "macie2.us-east-1.amazonaws.com",
	"managedblockchain":                     "managedblockchain.us-east-1.amazonaws.com",
	"marketplacecommerceanalytics":          "marketplacecommerceanalytics.us-east-1.amazonaws.com",
	"mediaconnect":                          "mediaconnect.us-east-1.amazonaws.com",
	"mediaconvert":                          "mediaconvert.us-east-1.amazonaws.com",
	"medialive":                             "medialive.us-east-1.amazonaws.com",
	"mediapackage":                          "mediapackage.us-east-1.amazonaws.com",
	"mediapackage-vod":                      "mediapackage-vod.us-east-1.amazonaws.com",
	"mediastore":                            "mediastore.us-east-1.amazonaws.com",
	"mediatailor":                           "api.mediatailor.us-east-1.amazonaws.com",
	"memorydb":                              "memory-db.us-east-1.amazonaws.com",
	"mgh":                                   "mgh.us-east-1.amazonaws.com",
	"mgn":                                   "mgn.us-east-1.amazonaws.com",
	"mobiletargeting":                       "pinpoint.us-east-1.amazonaws.com",
	"monitoring":                            "monitoring.us-east-1.amazonaws.com",
	"mq":                                    "mq.us-east-1.amazonaws.com",
	"mturk-requester":                       "mturk-requester.us-east-1.amazonaws.com",
	"network-firewall":                      "network-firewall.us-east-1.amazonaws.com",
	"networkmanager":                        "networkmanager.us-west-2.amazonaws.com", // Maps to us-west-2.
	"nimble":                                "nimble.us-east-1.amazonaws.com",
	"opsworks":                              "opsworks.us-east-1.amazonaws.com",
	"opsworks-cm":                           "opsworks-cm.us-east-1.amazonaws.com",
	"organizations":                         "organizations.us-east-1.amazonaws.com",
	"outposts":                              "outposts.us-east-1.amazonaws.com",
	"panorama":                              "panorama.us-east-1.amazonaws.com",
	"personalize":                           "personalize.us-east-1.amazonaws.com",
	"pi":                                    "pi.us-east-1.amazonaws.com",
	"polly":                                 "polly.us-east-1.amazonaws.com",
	"pricing":                               "api.pricing.us-east-1.amazonaws.com",
	"profile":                               "profile.us-east-1.amazonaws.com",
	"qldb":                                  "qldb.us-east-1.amazonaws.com",
	"quicksight":                            "quicksight.us-east-1.amazonaws.com",
	"ram":                                   "ram.us-east-1.amazonaws.com",
	"rbin":                                  "rbin.us-east-1.amazonaws.com",
	"rds":                                   "rds.us-east-1.amazonaws.com",
	"rds-data":                              "rds-data.us-east-1.amazonaws.com",
	"redshift":                              "redshift.us-east-1.amazonaws.com",
	"redshift-data":                         "redshift-data.us-east-1.amazonaws.com",
	"refactor-spaces":                       "refactor-spaces.us-east-1.amazonaws.com",
	"rekognition":                           "rekognition.us-east-1.amazonaws.com",
	"resiliencehub":                         "resiliencehub.us-east-1.amazonaws.com",
	"resource-groups":                       "resource-groups.us-east-1.amazonaws.com",
	"robomaker":                             "robomaker.us-east-1.amazonaws.com",
	"route53":                               "route53.amazonaws.com",
	"route53-recovery-cluster":              "route53-recovery-cluster.us-east-1.amazonaws.com",
	"route53-recovery-control-config":       "route53-recovery-control-config.us-east-1.amazonaws.com",
	"route53-recovery-readiness":            "route53-recovery-readiness.us-east-1.amazonaws.com",
	"route53domains":                        "route53domains.us-east-1.amazonaws.com",
	"route53resolver":                       "route53resolver.us-east-1.amazonaws.com",
	"rum":                                   "rum.us-east-1.amazonaws.com",
	"s3":                                    "s3.amazonaws.com",
	"s3-outposts":                           "s3-outposts.us-east-1.amazonaws.com",
	"sagemaker":                             "api.sagemaker.us-east-1.amazonaws.com",
	"savingsplans":                          "savingsplans.amazonaws.com",
	"schemas":                               "schemas.us-east-1.amazonaws.com",
	"secretsmanager":                        "secretsmanager.us-east-1.amazonaws.com",
	"securityhub":                           "securityhub.us-east-1.amazonaws.com",
	"serverlessrepo":                        "serverlessrepo.us-east-1.amazonaws.com",
	"servicecatalog":                        "servicecatalog.us-east-1.amazonaws.com",
	"servicediscovery":                      "servicediscovery.us-east-1.amazonaws.com",
	"servicequotas":                         "servicequotas.us-east-1.amazonaws.com",
	"ses":                                   "email.us-east-1.amazonaws.com",
	"shield":                                "shield.us-east-1.amazonaws.com",
	"signer":                                "signer.us-east-1.amazonaws.com",
	"sms":                                   "sms.us-east-1.amazonaws.com",
	"sms-voice":                             "sms-voice.pinpoint.us-east-1.amazonaws.com",
	"snow-device-management":                "snow-device-management.us-east-1.amazonaws.com",
	"snowball":                              "snowball.us-east-1.amazonaws.com",
	"sns":                                   "sns.us-east-1.amazonaws.com",
	"sqs":                                   "sqs.us-east-1.amazonaws.com",
	"ssm":                                   "ssm.us-east-1.amazonaws.com",
	"ssm-contacts":                          "ssm-contacts.us-east-1.amazonaws.com",
	"ssm-incidents":                         "ssm-incidents.us-east-1.amazonaws.com",
	"sso":                                   "sso.us-east-1.amazonaws.com",
	"states":                                "states.us-east-1.amazonaws.com",
	"storagegateway":                        "storagegateway.us-east-1.amazonaws.com",
	"sts":                                   "sts.amazonaws.com",
	"support":                               "support.us-east-1.amazonaws.com",
	"swf":                                   "swf.us-east-1.amazonaws.com",
	"synthetics":                            "synthetics.us-east-1.amazonaws.com",
	"tagging":                               "tagging.us-east-1.amazonaws.com",
	"textract":                              "textract.us-east-1.amazonaws.com",
	"timestream":                            "query.timestream.us-east-1.amazonaws.com",
	"transcribe":                            "transcribe.us-east-1.amazonaws.com",
	"transfer":                              "transfer.us-east-1.amazonaws.com",
	"translate":                             "translate.us-east-1.amazonaws.com",
	"voiceid":                               "voiceid.us-east-1.amazonaws.com",
	"waf":                                   "waf.amazonaws.com",
	"waf-regional":                          "waf-regional.us-east-1.amazonaws.com",
	"wafv2":                                 "wafv2.us-east-1.amazonaws.com",
	"wellarchitected":                       "wellarchitected.us-east-1.amazonaws.com",
	"wisdom":                                "wisdom.us-east-1.amazonaws.com",
	"workdocs":                              "workdocs.us-east-1.amazonaws.com",
	"worklink":                              "worklink.us-east-1.amazonaws.com",
	"workmail":                              "workmail.us-east-1.amazonaws.com",
	"workmailmessageflow":                   "workmailmessageflow.us-east-1.amazonaws.com",
	"workspaces":                            "workspaces.us-east-1.amazonaws.com",
	"workspaces-web":                        "workspaces-web.us-east-1.amazonaws.com",
	"xray":                                  "xray.us-east-1.amazonaws.com",

	// TODO here is a list of hostnames sharing same signing names. They are
	// currently commented out since we don't know how to handle them yet.
	// "apigateway":      "apigateway.us-east-1.amazonaws.com",
	// "apigateway":      "execute-api.us-east-1.amazonaws.com",
	// "aws-marketplace": "catalog.marketplace.us-east-1.amazonaws.com",
	// "aws-marketplace": "entitlement.marketplace.us-east-1.amazonaws.com",
	// "aws-marketplace": "metering.marketplace.us-east-1.amazonaws.com",
	// "chime":           "chime.us-east-1.amazonaws.com",
	// "chime":           "identity-chime.us-east-1.amazonaws.com",
	// "chime":           "meetings-chime.us-east-1.amazonaws.com",
	// "chime":           "messaging-chime.us-east-1.amazonaws.com",
	// "cloudhsm":        "cloudhsm.us-east-1.amazonaws.com",
	// "cloudhsm":        "cloudhsmv2.us-east-1.amazonaws.com",
	// "cloudsearch":     "cloudsearch.us-east-1.amazonaws.com",
	// "cloudsearch":     "cloudsearchdomain.us-east-1.amazonaws.com",
	// "connect":         "connect.us-east-1.amazonaws.com",
	// "connect":         "contact-lens.us-east-1.amazonaws.com",
	// "connect":         "participant.connect.us-east-1.amazonaws.com",
	// "dynamodb":        "dynamodb.us-east-1.amazonaws.com",
	// "dynamodb":        "streams.dynamodb.us-east-1.amazonaws.com",
	// "forecast":        "forecast.us-east-1.amazonaws.com",
	// "forecast":        "forecastquery.us-east-1.amazonaws.com",
	// "iot1click":       "devices.iot1click.us-east-1.amazonaws.com",
	// "iot1click":       "projects.iot1click.us-east-1.amazonaws.com",
	// "lex":             "models-v2-lex.us-east-1.amazonaws.com",
	// "lex":             "models.lex.us-east-1.amazonaws.com",
	// "lex":             "runtime-v2-lex.us-east-1.amazonaws.com",
	// "lex":             "runtime.lex.us-east-1.amazonaws.com",
	// "mediastore":      "data.mediastore.us-east-1.amazonaws.com",
	// "mediastore":      "mediastore.us-east-1.amazonaws.com",
	// "mgh":             "mgh.us-east-1.amazonaws.com",
	// "mgh":             "migrationhub-config.us-east-1.amazonaws.com",
	// "personalize":     "personalize-events.us-east-1.amazonaws.com",
	// "personalize":     "personalize-runtime.us-east-1.amazonaws.com",
	// "personalize":     "personalize.us-east-1.amazonaws.com",
	// "qldb":            "session.qldb.us-east-1.amazonaws.com",
	// "qldb":            "qldb.us-east-1.amazonaws.com",
	// "s3":              "s3-control.dualstack.us-east-1.amazonaws.com",
	// "s3":              "s3.amazonaws.com",
	// "sagemaker":       "a2i-runtime.sagemaker.us-east-1.amazonaws.com",
	// "sagemaker":       "api.sagemaker.us-east-1.amazonaws.com",
	// "sagemaker":       "edge.sagemaker.us-east-1.amazonaws.com",
	// "sagemaker":       "featurestore-runtime.sagemaker.us-east-1.amazonaws.com",
	// "sagemaker":       "runtime.sagemaker.us-east-1.amazonaws.com",
	// "servicecatalog":  "servicecatalog-appregistry.us-east-1.amazonaws.com",
	// "servicecatalog":  "servicecatalog.us-east-1.amazonaws.com",
	// "timestream":      "ingest.timestream.us-east-1.amazonaws.com",
	// "timestream":      "query.timestream.us-east-1.amazonaws.com",
	// "transcribe":      "transcribe.us-east-1.amazonaws.com",
	// "transcribe":      "transcribestreaming.us-east-1.amazonaws.com",
}

func TestResolveEndpoints(t *testing.T) {
	signer := v4.NewSigner(credentials.NewStaticCredentials("fakeClientKeyID", "fakeClientSecret", ""))
	region := "us-east-1"
	now := time.Now()

	for signingName := range signingNameToHostname {
		req, err := http.NewRequest("GET", "http://localhost", nil)
		require.NoError(t, err)

		_, err = signer.Sign(req, bytes.NewReader(nil), signingName, region, now)
		require.NoError(t, err)

		endpoint, err := resolveEndpoint(req)
		require.NoError(t, err)
		require.Equal(t, signingName, endpoint.SigningName)
		require.Equal(t, "https://"+signingNameToHostname[signingName], endpoint.URL, "for signing name %q", signingName)
	}

	t.Run("X-Forwarded-Host", func(t *testing.T) {
		req, err := http.NewRequest("GET", "http://localhost", nil)
		require.NoError(t, err)
		req.Header.Set("X-Forwarded-Host", "some-service.us-east-1.amazonaws.com")

		_, err = signer.Sign(req, bytes.NewReader(nil), "some-service", region, now)
		require.NoError(t, err)

		endpoint, err := resolveEndpoint(req)
		require.NoError(t, err)
		require.Equal(t, "some-service", endpoint.SigningName)
		require.Equal(t, "https://some-service.us-east-1.amazonaws.com", endpoint.URL)
	})
}
