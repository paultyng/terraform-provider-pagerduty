package pagerduty

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/heimweh/go-pagerduty/pagerduty"
)

const AppBaseUrl = "https://app.pagerduty.com"

func resourcePagerDutySlackConnection() *schema.Resource {
	return &schema.Resource{
		Create: resourcePagerDutySlackConnectionCreate,
		Read:   resourcePagerDutySlackConnectionRead,
		Update: resourcePagerDutySlackConnectionUpdate,
		Delete: resourcePagerDutySlackConnectionDelete,
		Importer: &schema.ResourceImporter{
			State: resourcePagerDutySlackConnectionImport,
		},
		Schema: map[string]*schema.Schema{
			"source_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"source_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"source_type": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validateValueFunc([]string{
					"service_reference",
					"team_reference",
				}),
			},
			"channel_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"channel_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"workspace_id": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("SLACK_CONNECTION_WORKSPACE_ID", nil),
			},
			"notification_type": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validateValueFunc([]string{
					"responder",
					"stakeholder",
				}),
			},
			"config": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"events": {
							Type:     schema.TypeList,
							Required: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"priorities": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"urgency": {
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: validateValueFunc([]string{
								"high",
								"low",
							}),
						},
					},
				},
			},
		},
	}
}

func buildSlackConnectionStruct(d *schema.ResourceData) (*pagerduty.SlackConnection, error) {
	slackConn := pagerduty.SlackConnection{
		SourceID:         d.Get("source_id").(string),
		SourceName:       d.Get("source_name").(string),
		SourceType:       d.Get("source_type").(string),
		ChannelID:        d.Get("channel_id").(string),
		ChannelName:      d.Get("channel_name").(string),
		WorkspaceID:      d.Get("workspace_id").(string),
		NotificationType: d.Get("notification_type").(string),
		Config:           expandConnectionConfig(d.Get("config").(interface{})),
	}
	return &slackConn, nil
}

func resourcePagerDutySlackConnectionCreate(d *schema.ResourceData, meta interface{}) error {
	config := setClientConfig()
	client, err := pagerduty.NewClient(config)
	if err != nil {
		return err
	}

	retryErr := resource.Retry(2*time.Minute, func() *resource.RetryError {

		slackConn, err := buildSlackConnectionStruct(d)
		if err != nil {
			return resource.NonRetryableError(err)
		}
		log.Printf("[INFO] Creating PagerDuty slack connection for source %s and slack channel %s", slackConn.SourceID, slackConn.ChannelID)

		if slackConn, _, err = client.SlackConnections.Create(slackConn.WorkspaceID, slackConn); err != nil {
			return resource.RetryableError(err)
		} else if slackConn != nil {
			d.SetId(slackConn.ID)
			d.Set("workspace_id", slackConn.WorkspaceID)
		}
		return nil
	})
	if retryErr != nil {
		time.Sleep(2 * time.Second)
		return retryErr
	}
	return resourcePagerDutySlackConnectionRead(d, meta)
}

func resourcePagerDutySlackConnectionRead(d *schema.ResourceData, meta interface{}) error {
	config := setClientConfig()
	client, err := pagerduty.NewClient(config)
	if err != nil {
		return err
	}

	log.Printf("[INFO] Reading PagerDuty slack connection %s", d.Id())

	workspaceID := d.Get("workspace_id").(string)
	log.Printf("[DEBUG] Read Slack Connection: workspace_id %s", workspaceID)

	retryErr := resource.Retry(2*time.Minute, func() *resource.RetryError {
		if slackConn, _, err := client.SlackConnections.Get(workspaceID, d.Id()); err != nil {
			return resource.RetryableError(err)
		} else if slackConn != nil {
			d.Set("source_id", slackConn.SourceID)
			d.Set("source_name", slackConn.SourceName)
			d.Set("source_type", slackConn.SourceType)
			d.Set("channel_id", slackConn.ChannelID)
			d.Set("channel_name", slackConn.ChannelName)
			d.Set("notification_type", slackConn.NotificationType)
			d.Set("config", flattenConnectionConfig(slackConn.Config))
		}
		return nil
	})

	if retryErr != nil {
		time.Sleep(2 * time.Second)
		return retryErr
	}

	return nil
}

func resourcePagerDutySlackConnectionUpdate(d *schema.ResourceData, meta interface{}) error {
	config := &pagerduty.Config{
		Token:   os.Getenv("PAGERDUTY_USER_TOKEN"),
		BaseURL: AppBaseUrl,
	}
	client, err := pagerduty.NewClient(config)
	if err != nil {
		return err
	}

	slackConn, err := buildSlackConnectionStruct(d)
	if err != nil {
		return err
	}
	log.Printf("[INFO] Updating PagerDuty slack connection %s", d.Id())

	if _, _, err := client.SlackConnections.Update(slackConn.WorkspaceID, d.Id(), slackConn); err != nil {
		return err
	}

	return nil
}

func resourcePagerDutySlackConnectionDelete(d *schema.ResourceData, meta interface{}) error {
	config := setClientConfig()
	client, err := pagerduty.NewClient(config)
	if err != nil {
		return err
	}

	log.Printf("[INFO] Deleting PagerDuty slack connection %s", d.Id())
	workspaceID := d.Get("workspace_id").(string)

	if _, err := client.SlackConnections.Delete(workspaceID, d.Id()); err != nil {
		return err
	}

	d.SetId("")

	return nil
}

func expandConnectionConfig(v interface{}) pagerduty.ConnectionConfig {
	c := v.([]interface{})[0].(map[string]interface{})

	var config pagerduty.ConnectionConfig

	config = pagerduty.ConnectionConfig{
		Events:     expandConfigList(c["events"].([]interface{})),
		Priorities: expandConfigList(c["priorities"].([]interface{})),
		Urgency:    nil,
	}
	if val, ok := c["urgency"]; ok {
		urgency := val.(string)
		if len(urgency) > 0 {
			config.Urgency = &urgency
		}
	}
	return config
}

func expandConfigList(v interface{}) []string {
	var items []string
	for _, i := range v.([]interface{}) {
		items = append(items, i.(string))
	}
	return items
}

func flattenConnectionConfig(config pagerduty.ConnectionConfig) []map[string]interface{} {
	var configs []map[string]interface{}
	configMap := map[string]interface{}{
		"events":     flattenConfigList(config.Events),
		"priorities": flattenConfigList(config.Priorities),
	}
	if config.Urgency != nil {
		configMap["urgency"] = *config.Urgency
	}
	configs = append(configs, configMap)
	return configs
}

func flattenConfigList(list []string) interface{} {
	var items []interface{}

	for _, i := range list {
		items = append(items, i)
	}

	return items
}

func setClientConfig() *pagerduty.Config {
	return &pagerduty.Config{
		Token:   os.Getenv("PAGERDUTY_USER_TOKEN"),
		BaseURL: AppBaseUrl,
	}
}

func resourcePagerDutySlackConnectionImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	config := setClientConfig()
	client, err := pagerduty.NewClient(config)
	if err != nil {
		return nil, err
	}

	ids := strings.Split(d.Id(), ".")

	if len(ids) != 2 {
		return []*schema.ResourceData{}, fmt.Errorf("Error importing pagerduty_slack_connection. Expecting an importation ID formed as '<workspace_id>.<slack_connection_id>'")
	}
	workspaceID, connectionID := ids[0], ids[1]

	_, _, err = client.SlackConnections.Get(workspaceID, connectionID)
	if err != nil {
		return []*schema.ResourceData{}, err
	}

	d.SetId(connectionID)
	d.Set("workspace_id", workspaceID)

	return []*schema.ResourceData{d}, nil
}
