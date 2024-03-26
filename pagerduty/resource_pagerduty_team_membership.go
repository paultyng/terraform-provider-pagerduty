package pagerduty

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/heimweh/go-pagerduty/pagerduty"
)

// Deprecated: Migrated to pagerdutyplugin.resourceTeamMembership. Kept for testing purposes.
func resourcePagerDutyTeamMembership() *schema.Resource {
	return &schema.Resource{
		Create: resourcePagerDutyTeamMembershipCreate,
		Read:   resourcePagerDutyTeamMembershipRead,
		Update: resourcePagerDutyTeamMembershipUpdate,
		Delete: resourcePagerDutyTeamMembershipDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"user_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"team_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"role": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "manager",
				ValidateDiagFunc: validateValueDiagFunc([]string{
					"observer",
					"responder",
					"manager",
				}),
			},
		},
	}
}

func maxRetries() int {
	return 4
}

func retryDelayMs() int {
	return 500
}

func calculateDelay(retryCount int) time.Duration {
	return time.Duration(retryCount*retryDelayMs()) * time.Millisecond
}

func fetchPagerDutyTeamMembershipWithRetries(d *schema.ResourceData, meta interface{}, errCallback func(error, *schema.ResourceData) error, retryCount int, neededRole string) error {
	if retryCount >= maxRetries() {
		return nil
	}
	if err := fetchPagerDutyTeamMembership(d, meta, errCallback); err != nil {
		return err
	}
	fetchedRole, userId, teamId := d.Get("role").(string), d.Get("user_id"), d.Get("team_id")
	if strings.Compare(neededRole, fetchedRole) == 0 {
		return nil
	}
	log.Printf("[DEBUG] Warning role '%s' fetched from PD is different from the role '%s' from config for user: %s from team: %s, retrying...", fetchedRole, neededRole, userId, teamId)

	retryCount++
	time.Sleep(calculateDelay(retryCount))
	return fetchPagerDutyTeamMembershipWithRetries(d, meta, errCallback, retryCount, neededRole)
}

func fetchPagerDutyTeamMembership(d *schema.ResourceData, meta interface{}, errCallback func(error, *schema.ResourceData) error) error {
	client, err := meta.(*Config).Client()
	if err != nil {
		return err
	}

	userID, teamID, err := resourcePagerDutyParseColonCompoundID(d.Id())
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Reading user: %s from team: %s", userID, teamID)
	return retry.Retry(2*time.Minute, func() *retry.RetryError {
		resp, _, err := client.Teams.GetMembers(teamID, &pagerduty.GetMembersOptions{})
		if err != nil {
			if isErrCode(err, http.StatusBadRequest) {
				return retry.NonRetryableError(err)
			}

			errResp := errCallback(err, d)
			if errResp != nil {
				time.Sleep(2 * time.Second)
				return retry.RetryableError(errResp)
			}

			return nil
		}

		for _, member := range resp.Members {
			if member.User.ID == userID {
				d.Set("user_id", userID)
				d.Set("team_id", teamID)
				d.Set("role", member.Role)

				return nil
			}
		}

		log.Printf("[WARN] Removing %s since the user: %s is not a member of: %s", d.Id(), userID, teamID)
		d.SetId("")

		return nil
	})
}

func resourcePagerDutyTeamMembershipCreate(d *schema.ResourceData, meta interface{}) error {
	client, err := meta.(*Config).Client()
	if err != nil {
		return err
	}

	userID := d.Get("user_id").(string)
	teamID := d.Get("team_id").(string)
	role := d.Get("role").(string)

	log.Printf("[DEBUG] Adding user: %s to team: %s with role: %s", userID, teamID, role)

	retryErr := retry.Retry(2*time.Minute, func() *retry.RetryError {
		if _, err := client.Teams.AddUserWithRole(teamID, userID, role); err != nil {
			if isErrCode(err, 500) {
				return retry.RetryableError(err)
			}

			return retry.NonRetryableError(err)
		}

		return nil
	})
	if retryErr != nil {
		return retryErr
	}

	d.SetId(fmt.Sprintf("%s:%s", userID, teamID))

	return fetchPagerDutyTeamMembershipWithRetries(d, meta, genError, 0, d.Get("role").(string))
}

func resourcePagerDutyTeamMembershipRead(d *schema.ResourceData, meta interface{}) error {
	return fetchPagerDutyTeamMembership(d, meta, handleNotFoundError)
}

func resourcePagerDutyTeamMembershipUpdate(d *schema.ResourceData, meta interface{}) error {
	client, err := meta.(*Config).Client()
	if err != nil {
		return err
	}

	userID := d.Get("user_id").(string)
	teamID := d.Get("team_id").(string)
	role := d.Get("role").(string)

	log.Printf("[DEBUG] Updating user: %s to team: %s with role: %s", userID, teamID, role)

	// To update existing membership resource, We can use the same API as creating a new membership.
	retryErr := retry.Retry(2*time.Minute, func() *retry.RetryError {
		if _, err := client.Teams.AddUserWithRole(teamID, userID, role); err != nil {
			if isErrCode(err, 500) {
				return retry.RetryableError(err)
			}

			return retry.NonRetryableError(err)
		}

		return nil
	})
	if retryErr != nil {
		return retryErr
	}

	d.SetId(fmt.Sprintf("%s:%s", userID, teamID))

	return fetchPagerDutyTeamMembershipWithRetries(d, meta, genError, 0, d.Get("role").(string))
}

func resourcePagerDutyTeamMembershipDelete(d *schema.ResourceData, meta interface{}) error {
	client, err := meta.(*Config).Client()
	if err != nil {
		return err
	}

	userID, teamID, err := resourcePagerDutyParseColonCompoundID(d.Id())
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Removing user: %s from team: %s", userID, teamID)

	// Extracting Escalation Policies ids where this team referenced
	epsAssociatedToUser, err := extractEPsAssociatedToUser(client, userID)
	if err != nil {
		return err
	}

	epsDissociatedFromTeam, err := dissociateEPsFromTeam(client, teamID, epsAssociatedToUser)
	if err != nil {
		return err
	}

	// Retrying to give other resources (such as escalation policies) to delete
	retryErr := retry.Retry(2*time.Minute, func() *retry.RetryError {
		if _, err := client.Teams.RemoveUser(teamID, userID); err != nil {
			if isErrCode(err, 400) {
				return retry.RetryableError(err)
			}

			return retry.NonRetryableError(err)
		}
		return nil
	})
	if retryErr != nil {
		time.Sleep(2 * time.Second)
		return retryErr
	}

	d.SetId("")

	err = associateEPsBackToTeam(client, teamID, epsDissociatedFromTeam)
	if err != nil {
		return err
	}

	return nil
}

func buildEPsIdsList(l []*pagerduty.OnCall) []string {
	eps := []string{}
	for _, o := range l {
		if o.EscalationPolicy != nil {
			eps = append(eps, o.EscalationPolicy.ID)
		}
	}
	return unique(eps)
}

func extractEPsAssociatedToUser(c *pagerduty.Client, userID string) ([]string, error) {
	var oncalls []*pagerduty.OnCall
	retryErr := retry.Retry(2*time.Minute, func() *retry.RetryError {
		resp, _, err := c.OnCall.List(&pagerduty.ListOnCallOptions{UserIds: []string{userID}})
		if err != nil {
			if isErrCode(err, http.StatusBadRequest) {
				return retry.NonRetryableError(err)
			}

			time.Sleep(2 * time.Second)
			return retry.RetryableError(err)
		}
		oncalls = resp.Oncalls
		return nil
	})
	if retryErr != nil {
		return nil, retryErr
	}
	epsAssociatedToUser := buildEPsIdsList(oncalls)
	return epsAssociatedToUser, nil
}

func dissociateEPsFromTeam(c *pagerduty.Client, teamID string, eps []string) ([]string, error) {
	epsDissociatedFromTeam := []string{}
	for _, ep := range eps {
		retryErr := retry.Retry(2*time.Minute, func() *retry.RetryError {
			_, err := c.Teams.RemoveEscalationPolicy(teamID, ep)
			if err != nil && !isErrCode(err, 404) {
				time.Sleep(2 * time.Second)
				return retry.RetryableError(err)
			}
			if err != nil && isErrCode(err, 404) {
				return retry.NonRetryableError(err)
			}
			return nil
		})
		if retryErr != nil {
			if !isErrCode(retryErr, 404) {
				return nil, fmt.Errorf("%w; Error while trying to dissociate Team %q from Escalation Policy %q", retryErr, teamID, ep)
			} else {
				// Skip Escaltion Policies not found. This happens when a destroy
				// operation is requested and Escalation Policy is destroyed first.
				continue
			}
		}
		epsDissociatedFromTeam = append(epsDissociatedFromTeam, ep)
		log.Printf("[DEBUG] EscalationPolicy %s removed from team %s", ep, teamID)
	}
	return epsDissociatedFromTeam, nil
}

func associateEPsBackToTeam(c *pagerduty.Client, teamID string, eps []string) error {
	for _, ep := range eps {
		retryErr := retry.Retry(2*time.Minute, func() *retry.RetryError {
			_, err := c.Teams.AddEscalationPolicy(teamID, ep)
			if err != nil && !isErrCode(err, 404) {
				time.Sleep(2 * time.Second)
				return retry.RetryableError(err)
			}
			return nil
		})
		if retryErr != nil {
			if !isErrCode(retryErr, 404) {
				return fmt.Errorf("%w; Error while trying to associate back team %q to Escalation Policy %q. Resource succesfully deleted, but some team association couldn't be completed, so you need to run \"terraform plan -refresh-only\" and again \"terraform apply/destroy\" in order to remediate the drift", retryErr, teamID, ep)
			} else {
				// Skip Escaltion Policies not found. This happens when a destroy
				// operation is requested and Escalation Policy is destroyed first.
				continue
			}
		}
		log.Printf("[DEBUG] EscalationPolicy %s added to team %s", ep, teamID)
	}
	return nil
}

func isTeamMember(user *pagerduty.User, teamID string) bool {
	var found bool

	for _, team := range user.Teams {
		if teamID == team.ID {
			found = true
		}
	}

	return found
}
