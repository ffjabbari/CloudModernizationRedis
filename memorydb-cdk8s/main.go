package main

import (
	"log"
	"os"
	"strings"

	"example.com/memorydb-cdk8s/imports/acl_memorydbservicesk8saws"
	"example.com/memorydb-cdk8s/imports/memorydbservicesk8saws"
	"example.com/memorydb-cdk8s/imports/servicesk8saws"
	"example.com/memorydb-cdk8s/imports/subnetgroups_memorydbservicesk8saws"
	"example.com/memorydb-cdk8s/imports/users_memorydbservicesk8saws"

	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/cdk8s-team/cdk8s-plus-go/cdk8splus22/v2"
)

type MyChartProps struct {
	cdk8s.ChartProps
}

const appPort = 8080
const lbPort = 9090

// field export
const fieldExportNameForClusterEndpoint = "export-memorydb-endpoint"
const fieldExportNameForUsername = "export-memorydb-username"
const configMapName = "export-memorydb-info"

var cfgMap cdk8splus22.ConfigMap

var fieldExportForClusterEndpoint servicesk8saws.FieldExport
var fieldExportForUsername servicesk8saws.FieldExport

var secret cdk8splus22.Secret
var memoryDBCluster memorydbservicesk8saws.Cluster
var user users_memorydbservicesk8saws.User

const secretName = "memdb-secret"
const secretKeyName = "password"

var memoryDBClusterName string
var memoryDBPassword string
var subnetIDs string
var securityGroupID string

const memoryDBUserAccessString = "on ~* &* +@all"

var memoryDBUsername string
var appDockerImage string

const memoryDBACLName = "demo-acl"
const memoryDBSubnetGroup = "demo-subnet-group"

const memoryDBNodeType = "db.t4g.small"
const memoryDBEngineVersion = "6.2"

func init() {

	memoryDBClusterName = os.Getenv("MEMORYDB_CLUSTER_NAME")
	if memoryDBClusterName == "" {
		memoryDBClusterName = "memorydb-cluster-ack-cdk8s"
		log.Println("using default cluster name for memorydb - memorydb-cluster-ack-cdk8s")
	}

	memoryDBUsername = os.Getenv("MEMORYDB_USERNAME")
	if memoryDBUsername == "" {
		memoryDBUsername = "demouser"
		log.Println("using default username for memorydb - demouser")
	}

	memoryDBPassword = os.Getenv("MEMORYDB_PASSWORD")
	if memoryDBPassword == "" {
		memoryDBPassword = "Password123456789"
		log.Println("using default password for memorydb - Password123456789")
	}

	subnetIDs = os.Getenv("SUBNET_ID_LIST")
	if subnetIDs == "" {
		log.Fatal("missing environment variable SUBNET_ID_LIST")
	}

	securityGroupID = os.Getenv("SECURITY_GROUP_ID")
	if securityGroupID == "" {
		log.Fatal("missing environment variable SECURITY_GROUP_ID")
	}

	appDockerImage = os.Getenv("DOCKER_IMAGE")
	if appDockerImage == "" {
		log.Fatal("missing environment variable DOCKER_IMAGE")
	}

}

func NewMemoryDBChart(scope constructs.Construct, id string, props *MyChartProps) cdk8s.Chart {
	var cprops cdk8s.ChartProps
	if props != nil {
		cprops = props.ChartProps
	}
	chart := cdk8s.NewChart(scope, jsii.String(id), &cprops)

	secret = cdk8splus22.NewSecret(chart, jsii.String("password"), &cdk8splus22.SecretProps{
		Metadata:   &cdk8s.ApiObjectMetadata{Name: jsii.String(secretName)},
		StringData: &map[string]*string{"password": jsii.String(memoryDBPassword)},
	})

	user = users_memorydbservicesk8saws.NewUser(chart, jsii.String("user"), &users_memorydbservicesk8saws.UserProps{
		Metadata: &cdk8s.ApiObjectMetadata{Name: jsii.String(memoryDBUsername)},
		Spec: &users_memorydbservicesk8saws.UserSpec{
			Name:         jsii.String(memoryDBUsername),
			AccessString: jsii.String(memoryDBUserAccessString),
			AuthenticationMode: &users_memorydbservicesk8saws.UserSpecAuthenticationMode{
				Type: jsii.String("Password"),
				Passwords: &[]*users_memorydbservicesk8saws.UserSpecAuthenticationModePasswords{
					{Name: secret.Name(), Key: jsii.String(secretKeyName)},
				},
			},
		},
	})

	acl := acl_memorydbservicesk8saws.NewAcl(chart, jsii.String("acl"),
		&acl_memorydbservicesk8saws.AclProps{
			Metadata: &cdk8s.ApiObjectMetadata{Name: jsii.String(memoryDBACLName)},
			Spec: &acl_memorydbservicesk8saws.AclSpec{
				Name:      jsii.String(memoryDBACLName),
				UserNames: jsii.Strings(*user.Name()),
			},
		})

	subnetGroup := subnetgroups_memorydbservicesk8saws.NewSubnetGroup(chart, jsii.String("sg"),
		&subnetgroups_memorydbservicesk8saws.SubnetGroupProps{

			Metadata: &cdk8s.ApiObjectMetadata{Name: jsii.String(memoryDBSubnetGroup)},
			Spec: &subnetgroups_memorydbservicesk8saws.SubnetGroupSpec{
				Name: jsii.String(memoryDBSubnetGroup),
				//SubnetIDs: jsii.Strings("subnet-086c4a45ec9a206e1", "subnet-0d9a9c6d2ca7a24df", "subnet-028ca54bb859a4994"), //same as EKS clsuter
				SubnetIDs: jsii.Strings(strings.Split(subnetIDs, ",")...), //same as EKS clsuter
			},
		})

	memoryDBCluster = memorydbservicesk8saws.NewCluster(chart, jsii.String("memorydb-ack-cdk8s"),
		&memorydbservicesk8saws.ClusterProps{
			Metadata: &cdk8s.ApiObjectMetadata{Name: jsii.String(memoryDBClusterName)},
			Spec: &memorydbservicesk8saws.ClusterSpec{
				Name:                jsii.String(memoryDBClusterName),
				NodeType:            jsii.String(memoryDBNodeType),
				AclName:             acl.Name(),
				EngineVersion:       jsii.String(memoryDBEngineVersion),
				NumShards:           jsii.Number(1),
				NumReplicasPerShard: jsii.Number(1),
				SecurityGroupIDs:    jsii.Strings(securityGroupID), //same as EKS clsuter clusterSecurityGroupId
				SubnetGroupName:     subnetGroup.Name(),
			},
		})

	return chart
}

func NewConfigChart(scope constructs.Construct, id string, props *MyChartProps) cdk8s.Chart {

	var cprops cdk8s.ChartProps
	if props != nil {
		cprops = props.ChartProps
	}
	chart := cdk8s.NewChart(scope, jsii.String(id), &cprops)

	cfgMap = cdk8splus22.NewConfigMap(chart, jsii.String("config-map"),
		&cdk8splus22.ConfigMapProps{
			Metadata: &cdk8s.ApiObjectMetadata{
				Name: jsii.String(configMapName)}})

	fieldExportForClusterEndpoint = servicesk8saws.NewFieldExport(chart, jsii.String("fexp-cluster"), &servicesk8saws.FieldExportProps{
		Metadata: &cdk8s.ApiObjectMetadata{Name: jsii.String(fieldExportNameForClusterEndpoint)},
		Spec: &servicesk8saws.FieldExportSpec{
			From: &servicesk8saws.FieldExportSpecFrom{Path: jsii.String(".status.clusterEndpoint.address"),
				Resource: &servicesk8saws.FieldExportSpecFromResource{
					Group: jsii.String("memorydb.services.k8s.aws"),
					Kind:  jsii.String("Cluster"),
					Name:  memoryDBCluster.Name()}},
			To: &servicesk8saws.FieldExportSpecTo{
				Name: cfgMap.Name(),
				Kind: servicesk8saws.FieldExportSpecToKind_CONFIGMAP}}})

	fieldExportForUsername = servicesk8saws.NewFieldExport(chart, jsii.String("fexp-username"), &servicesk8saws.FieldExportProps{
		Metadata: &cdk8s.ApiObjectMetadata{Name: jsii.String(fieldExportNameForUsername)},
		Spec: &servicesk8saws.FieldExportSpec{
			From: &servicesk8saws.FieldExportSpecFrom{Path: jsii.String(".spec.name"),
				Resource: &servicesk8saws.FieldExportSpecFromResource{
					Group: jsii.String("memorydb.services.k8s.aws"),
					Kind:  jsii.String("User"),
					Name:  user.Name()}},
			To: &servicesk8saws.FieldExportSpecTo{
				Name: cfgMap.Name(),
				Kind: servicesk8saws.FieldExportSpecToKind_CONFIGMAP}}})

	return chart
}

func NewDeploymentChart(scope constructs.Construct, id string, props *MyChartProps) cdk8s.Chart {
	var cprops cdk8s.ChartProps
	if props != nil {
		cprops = props.ChartProps
	}
	chart := cdk8s.NewChart(scope, jsii.String(id), &cprops)

	dep := cdk8splus22.NewDeployment(chart, jsii.String("memorydb-app-deployment"), &cdk8splus22.DeploymentProps{
		Metadata: &cdk8s.ApiObjectMetadata{
			Name: jsii.String("memorydb-app")}})

	container := dep.AddContainer(
		&cdk8splus22.ContainerProps{
			Name:  jsii.String("memorydb-app-container"),
			Image: jsii.String(appDockerImage),
			Port:  jsii.Number(appPort)})

	container.Env().AddVariable(jsii.String("MEMORYDB_CLUSTER_ENDPOINT"),
		cdk8splus22.EnvValue_FromConfigMap(
			cfgMap,
			jsii.String("default."+*fieldExportForClusterEndpoint.Name()),
			&cdk8splus22.EnvValueFromConfigMapOptions{Optional: jsii.Bool(false)}))

	container.Env().AddVariable(jsii.String("MEMORYDB_USERNAME"),
		cdk8splus22.EnvValue_FromConfigMap(
			cfgMap,
			jsii.String("default."+*fieldExportForUsername.Name()),
			&cdk8splus22.EnvValueFromConfigMapOptions{Optional: jsii.Bool(false)}))

	container.Env().AddVariable(jsii.String("MEMORYDB_PASSWORD"),
		cdk8splus22.EnvValue_FromSecretValue(
			&cdk8splus22.SecretValue{
				Secret: secret,
				Key:    jsii.String("password")},
			&cdk8splus22.EnvValueFromSecretOptions{}))

	dep.ExposeViaService(
		&cdk8splus22.DeploymentExposeViaServiceOptions{
			Name:        jsii.String("memorydb-app-service"),
			ServiceType: cdk8splus22.ServiceType_LOAD_BALANCER,
			Ports: &[]*cdk8splus22.ServicePort{
				{Protocol: cdk8splus22.Protocol_TCP,
					Port:       jsii.Number(lbPort),
					TargetPort: jsii.Number(appPort)}}})

	return chart
}

func main() {
	app := cdk8s.NewApp(nil)

	memorydb := NewMemoryDBChart(app, "memorydb", nil)

	config := NewConfigChart(app, "config", nil)
	config.AddDependency(memorydb)

	deployment := NewDeploymentChart(app, "deployment", nil)
	deployment.AddDependency(memorydb, config)

	app.Synth()
}
