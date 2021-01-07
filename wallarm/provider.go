package wallarm

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"

	"github.com/416e64726579/terraform-provider-wallarm/version"
	wallarm "github.com/416e64726579/wallarm-go"

	cleanhttp "github.com/hashicorp/go-cleanhttp"
	"github.com/hashicorp/terraform-plugin-sdk/helper/logging"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/hashicorp/terraform-plugin-sdk/httpclient"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

// Provider returns a terraform.ResourceProvider
func Provider() terraform.ResourceProvider {
	provider := &schema.Provider{
		Schema: map[string]*schema.Schema{
			"api_host": {
				Type:         schema.TypeString,
				Optional:     true,
				DefaultFunc:  schema.EnvDefaultFunc("WALLARM_API_HOST", "https://api.wallarm.com"),
				Description:  "The API host address of the Wallarm Cloud for operations",
				ValidateFunc: validation.IsURLWithHTTPS,
			},

			"api_uuid": {
				Type:         schema.TypeString,
				Optional:     true,
				DefaultFunc:  schema.EnvDefaultFunc("WALLARM_API_UUID", nil),
				Description:  "The API UUID of the user for operations",
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`[0-9a-f\-]+`), "API key must only contain characters 0-9 and a-f (all lowercased)"),
				Sensitive:    true,
			},

			"api_secret": {
				Type:         schema.TypeString,
				Optional:     true,
				DefaultFunc:  schema.EnvDefaultFunc("WALLARM_API_SECRET", nil),
				Description:  "The API Secret of the user for operations",
				ValidateFunc: validation.StringMatch(regexp.MustCompile("[A-Za-z0-9-_]{40}"), "API tokens must only contain characters a-z, A-Z, 0-9 and underscores"),
				Sensitive:    true,
			},

			"client_id": {
				Type:        schema.TypeInt,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("WALLARM_API_CLIENT_ID", nil),
				Description: "The Client ID to perform changes on",
			},

			"retries": {
				Type:        schema.TypeInt,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("WALLARM_API_RETRIES", 3),
				Description: "Maximum number of retries to perform when an API request fails",
			},

			"min_backoff": {
				Type:        schema.TypeInt,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("WALLARM_API_MIN_BACKOFF", 1),
				Description: "Minimum backoff period in seconds after failed API calls",
			},

			"max_backoff": {
				Type:        schema.TypeInt,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("WALLARM_API_MAX_BACKOFF", 5),
				Description: "Maximum backoff period in seconds after failed API calls",
			},

			"api_client_logging": {
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("WALLARM_API_CLIENT_LOGGING", false),
				Description: "Whether to print logs from the API client (using the default log library logger)",
			},

			"ignore_existing": {
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("WALLARM_IGNORE_EXISTING_RESOURCES", false),
				Description: "Whether ignore or raise an exception when a resource exists.",
			},
		},

		DataSourcesMap: map[string]*schema.Resource{
			"wallarm_node": dataSourceWallarmNode(),
			"wallarm_vuln": dataSourceWallarmVuln(),
		},

		ResourcesMap: map[string]*schema.Resource{
			"wallarm_user":                          resourceWallarmUser(),
			"wallarm_blacklist":                     resourceWallarmBlacklist(),
			"wallarm_global_mode":                   resourceWallarmGlobalMode(),
			"wallarm_node":                          resourceWallarmNode(),
			"wallarm_scanner":                       resourceWallarmScanner(),
			"wallarm_application":                   resourceWallarmApp(),
			"wallarm_integration_email":             resourceWallarmEmail(),
			"wallarm_integration_opsgenie":          resourceWallarmOpsGenie(),
			"wallarm_integration_slack":             resourceWallarmSlack(),
			"wallarm_integration_pagerduty":         resourceWallarmPagerDuty(),
			"wallarm_integration_sumologic":         resourceWallarmSumologic(),
			"wallarm_integration_insightconnect":    resourceWallarmInsightConnect(),
			"wallarm_integration_splunk":            resourceWallarmSplunk(),
			"wallarm_integration_webhook":           resourceWallarmWebhook(),
			"wallarm_trigger":                       resourceWallarmTrigger(),
			"wallarm_rule_vpatch":                   resourceWallarmVpatch(),
			"wallarm_rule_mode":                     resourceWallarmMode(),
			"wallarm_rule_masking":                  resourceWallarmSensitiveData(),
			"wallarm_rule_regex":                    resourceWallarmRegex(),
			"wallarm_rule_ignore_regex":             resourceWallarmIgnoreRegex(),
			"wallarm_rule_attack_rechecker":         resourceWallarmAttackRechecker(),
			"wallarm_rule_attack_rechecker_rewrite": resourceWallarmAttackRecheckerRewrite(),
			"wallarm_rule_set_response_header":      resourceWallarmSetResponseHeader(),
			"wallarm_rule_bruteforce_counter":       resourceWallarmBruteForceCounter(),
			"wallarm_rule_dirbust_counter":          resourceWallarmDirbustCounter(),
		},
	}

	provider.ConfigureFunc = func(d *schema.ResourceData) (interface{}, error) {
		terraformVersion := provider.TerraformVersion
		return providerConfigure(d, terraformVersion)
	}

	return provider
}

func providerConfigure(d *schema.ResourceData, terraformVersion string) (interface{}, error) {
	retryOpt := wallarm.UsingRetryPolicy(d.Get("retries").(int), d.Get("min_backoff").(int), d.Get("max_backoff").(int))
	options := []wallarm.Option{retryOpt}

	if d.Get("api_client_logging").(bool) {
		options = append(options, wallarm.UsingLogger(log.New(os.Stderr, "", log.LstdFlags|log.Lshortfile)))
	}

	// c := cleanhttp.DefaultClient()
	c := cleanhttp.DefaultPooledClient()
	c.Transport = logging.NewTransport("Wallarm", c.Transport)
	options = append(options, wallarm.HTTPClient(c))

	tfUserAgent := httpclient.TerraformUserAgent(terraformVersion)
	providerUserAgent := fmt.Sprintf("terraform-provider-wallarm")
	ua := fmt.Sprintf("%s/%s/%s", tfUserAgent, providerUserAgent, version.ProviderVersion)
	options = append(options, wallarm.UserAgent(ua))
	options = append(options, wallarm.UsingBaseURL(apiURL))

	authHeaders := make(http.Header)
	config := Config{Options: options}

	if v, ok := d.GetOk("api_uuid"); ok {
		config.apiUUID = v.(string)
		authHeaders.Add("X-WallarmAPI-UUID", v.(string))
	} else {
		return nil, wallarm.ErrInvalidCredentials
	}
	if v, ok := d.GetOk("api_secret"); ok {
		config.apiSecret = v.(string)
		authHeaders.Add("X-WallarmAPI-Secret", v.(string))
	} else {
		return nil, wallarm.ErrInvalidCredentials
	}

	if v, ok := d.GetOk("api_host"); ok {
		config.apiURL = v.(string)
	}
	options = append(options, wallarm.Headers(authHeaders))
	config.Options = options

	client, err := config.Client()
	if err != nil {
		return nil, err
	}

	if v, ok := d.GetOk("client_id"); ok {
		ClientID = v.(int)
	} else {
		u, err := client.UserDetails()
		if err != nil {
			return nil, err
		}
		ClientID = u.Body.Clientid
	}

	return client, err
}
