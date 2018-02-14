package cmd

import (
	"github.com/jpignata/fargate/console"
	ELBV2 "github.com/jpignata/fargate/elbv2"
	Route53 "github.com/jpignata/fargate/route53"
	"github.com/spf13/cobra"
)

type LbAliasOperation struct {
	LoadBalancerName string
	AliasDomain      string
}

var lbAliasCmd = &cobra.Command{
	Use:   "alias <load-balancer-name> <domain-name>",
	Args:  cobra.ExactArgs(2),
	Short: "Create a load balancer alias record",
	Long: `Create a load balancer alias record

Create an alias record to the load balancer for domains that are hosted within
Amazon Route 53 and within the same AWS account. If you're using another DNS
provider or host your domains in a differnt account, you will need to manually
create this record.  `,
	Run: func(cmd *cobra.Command, args []string) {
		operation := &LbAliasOperation{
			LoadBalancerName: args[0],
			AliasDomain:      args[1],
		}

		createAliasRecord(operation)
	},
}

func init() {
	lbCmd.AddCommand(lbAliasCmd)
}

func createAliasRecord(operation *LbAliasOperation) {
	route53 := Route53.New(sess)
	elbv2 := ELBV2.New(sess)

	hostedZones, err := route53.ListHostedZones()

	if err != nil {
		console.ErrorExit(err, "Could not alias load balancer")
		return
	}

	loadBalancer := elbv2.DescribeLoadBalancer(operation.LoadBalancerName)

	if hostedZone, ok := hostedZones.FindSuperDomainOf(operation.AliasDomain); ok {
		route53.CreateAlias(hostedZone, "A", operation.AliasDomain, loadBalancer.DNSName, loadBalancer.HostedZoneId)
		console.Info("Created alias record (%s -> %s)", operation.AliasDomain, loadBalancer.DNSName)

		return
	}

	console.Issue("Could not find hosted zone for %s", operation.AliasDomain)
	console.Info("If you're hosting this domain elsewhere or in another AWS account, please manually create the Alias record:")
	console.Info("%s -> %s", operation.AliasDomain, loadBalancer.DNSName)
}
