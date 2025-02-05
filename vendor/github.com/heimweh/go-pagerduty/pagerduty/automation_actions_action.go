package pagerduty

import "fmt"

// AutomationActionsAction handles the communication with Automation Actions
// related methods of the PagerDuty API.
type AutomationActionsActionService service

type AutomationActionsAction struct {
	ID                                 string                               `json:"id"`
	Name                               string                               `json:"name"`
	Description                        *string                              `json:"description,omitempty"`
	ActionType                         string                               `json:"action_type"`
	RunnerID                           *string                              `json:"runner,omitempty"`
	ActionDataReference                AutomationActionsActionDataReference `json:"action_data_reference"`
	Services                           []*ServiceReference                  `json:"services,omitempty"`
	Teams                              []*TeamReference                     `json:"teams,omitempty"`
	Privileges                         *AutomationActionsPrivileges         `json:"privileges,omitempty"`
	Type                               *string                              `json:"type,omitempty"`
	ActionClassification               *string                              `json:"action_classification,omitempty"`
	RunnerType                         *string                              `json:"runner_type,omitempty"`
	CreationTime                       *string                              `json:"creation_time,omitempty"`
	ModifyTime                         *string                              `json:"modify_time,omitempty"`
	OnlyInvocableOnUnresolvedIncidents *bool                                `json:"only_invocable_on_unresolved_incidents,omitempty"`
	MapToAllServices                   *bool                                `json:"map_to_all_services,omitempty"`
}

type AutomationActionsActionDataReference struct {
	ProcessAutomationJobId        *string `json:"process_automation_job_id,omitempty"`
	ProcessAutomationJobArguments *string `json:"process_automation_job_arguments,omitempty"`
	ProcessAutomationNodeFilter   *string `json:"process_automation_node_filter,omitempty"`
	Script                        *string `json:"script,omitempty"`
	InvocationCommand             *string `json:"invocation_command,omitempty"`
}

type AutomationActionsActionPayload struct {
	Action *AutomationActionsAction `json:"action,omitempty"`
}

type AutomationActionsActionTeamAssociationPayload struct {
	Team *TeamReference `json:"team,omitempty"`
}

type AutomationActionsActionServiceAssociationPayload struct {
	Service *ServiceReference `json:"service,omitempty"`
}

var automationActionsActionBaseUrl = "/automation_actions/actions"

// Create creates a new action
func (s *AutomationActionsActionService) Create(action *AutomationActionsAction) (*AutomationActionsAction, *Response, error) {
	u := automationActionsActionBaseUrl
	v := new(AutomationActionsActionPayload)

	resp, err := s.client.newRequestDoOptions("POST", u, nil, &AutomationActionsActionPayload{Action: action}, &v)
	if err != nil {
		return nil, nil, err
	}

	return v.Action, resp, nil
}

// Get retrieves information about an action.
func (s *AutomationActionsActionService) Get(id string) (*AutomationActionsAction, *Response, error) {
	u := fmt.Sprintf("%s/%s", automationActionsActionBaseUrl, id)
	v := new(AutomationActionsActionPayload)

	resp, err := s.client.newRequestDoOptions("GET", u, nil, nil, &v)
	if err != nil {
		return nil, nil, err
	}

	return v.Action, resp, nil
}

// Update an existing action
func (s *AutomationActionsActionService) Update(ID string, action *AutomationActionsAction) (*AutomationActionsAction, *Response, error) {
	u := fmt.Sprintf("%s/%s", automationActionsActionBaseUrl, ID)
	v := new(AutomationActionsActionPayload)
	p := &AutomationActionsActionPayload{Action: action}

	resp, err := s.client.newRequestDo("PUT", u, nil, p, v)
	if err != nil {
		return nil, nil, err
	}

	return v.Action, resp, nil
}

// Delete deletes an existing action.
func (s *AutomationActionsActionService) Delete(id string) (*Response, error) {
	u := fmt.Sprintf("%s/%s", automationActionsActionBaseUrl, id)

	return s.client.newRequestDoOptions("DELETE", u, nil, nil, nil)
}

// Associate an Automation Action with a team
func (s *AutomationActionsActionService) AssociateToTeam(actionID, teamID string) (*AutomationActionsActionTeamAssociationPayload, *Response, error) {
	u := fmt.Sprintf("%s/%s/teams", automationActionsActionBaseUrl, actionID)
	v := new(AutomationActionsActionTeamAssociationPayload)
	p := &AutomationActionsActionTeamAssociationPayload{
		Team: &TeamReference{ID: teamID, Type: "team_reference"},
	}

	resp, err := s.client.newRequestDoOptions("POST", u, nil, p, &v)
	if err != nil {
		return nil, nil, err
	}

	return v, resp, nil
}

// Dissociate an Automation Action with a team
func (s *AutomationActionsActionService) DissociateToTeam(actionID, teamID string) (*Response, error) {
	u := fmt.Sprintf("%s/%s/teams/%s", automationActionsActionBaseUrl, actionID, teamID)

	return s.client.newRequestDoOptions("DELETE", u, nil, nil, nil)
}

// Gets the details of an Automation Action / team relation
func (s *AutomationActionsActionService) GetAssociationToTeam(actionID, teamID string) (*AutomationActionsActionTeamAssociationPayload, *Response, error) {
	u := fmt.Sprintf("%s/%s/teams/%s", automationActionsActionBaseUrl, actionID, teamID)
	v := new(AutomationActionsActionTeamAssociationPayload)

	resp, err := s.client.newRequestDoOptions("GET", u, nil, nil, &v)
	if err != nil {
		return nil, nil, err
	}

	return v, resp, nil
}

// Associate an Automation Action with a service
func (s *AutomationActionsActionService) AssociateToService(actionID, serviceID string) (*AutomationActionsActionServiceAssociationPayload, *Response, error) {
	u := fmt.Sprintf("%s/%s/services", automationActionsActionBaseUrl, actionID)
	v := new(AutomationActionsActionServiceAssociationPayload)
	p := &AutomationActionsActionServiceAssociationPayload{
		Service: &ServiceReference{ID: serviceID, Type: "service_reference"},
	}

	resp, err := s.client.newRequestDoOptions("POST", u, nil, p, &v)
	if err != nil {
		return nil, nil, err
	}

	return v, resp, nil
}

// Dissociate an Automation Action with a service
func (s *AutomationActionsActionService) DissociateFromService(actionID, serviceID string) (*Response, error) {
	u := fmt.Sprintf("%s/%s/services/%s", automationActionsActionBaseUrl, actionID, serviceID)

	return s.client.newRequestDoOptions("DELETE", u, nil, nil, nil)
}

// Gets the details of an Automation Action / service relation
func (s *AutomationActionsActionService) GetAssociationToService(actionID, serviceID string) (*AutomationActionsActionServiceAssociationPayload, *Response, error) {
	u := fmt.Sprintf("%s/%s/services/%s", automationActionsActionBaseUrl, actionID, serviceID)
	v := new(AutomationActionsActionServiceAssociationPayload)

	resp, err := s.client.newRequestDoOptions("GET", u, nil, nil, &v)
	if err != nil {
		return nil, nil, err
	}

	return v, resp, nil
}
