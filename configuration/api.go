package configuration

import (
	"errors"
	"log"
	"net/http"

	"github.com/livechat/lc-sdk-go/v6/authorization"
	i "github.com/livechat/lc-sdk-go/v6/internal"
)

type configurationAPI interface {
	Call(string, interface{}, interface{}, ...*i.CallOptions) error
	SetCustomHost(string)
	SetRetryStrategy(i.RetryStrategyFunc)
	SetStatsSink(i.StatsSinkFunc)
	SetLogger(*log.Logger)
}

// API provides the API operation methods for making requests to Livechat Configuration API via Web API.
// See this package's package overview docs for details on the service.
type API struct {
	configurationAPI
}

type GroupProperties struct {
	ID         int32
	Properties Properties
}

// NewAPI returns ready to use Configuration API.
//
// If provided client is nil, then default http client with 20s timeout is used.
func NewAPI(t authorization.TokenGetter, client *http.Client, clientID string) (*API, error) {
	api, err := i.NewAPI(t, client, clientID, i.DefaultHTTPRequestGenerator("configuration"))
	if err != nil {
		return nil, err
	}
	return &API{api}, nil
}

// RegisterWebhook allows to register specified webhook.
//
// When authorizing via Personal Access Token, set correct ClientID in opts.
func (a *API) RegisterWebhook(webhook *Webhook, opts *ManageWebhooksDefinitionOptions) (string, error) {
	var resp registerWebhookResponse
	var clientID string
	if opts != nil {
		clientID = opts.ClientID
	}
	err := a.Call("register_webhook", &registerWebhookRequest{webhook, clientID}, &resp)

	return resp.ID, err
}

// ListWebhooks returns configurations of all registered webhooks for requester's clientID.
//
// When authorizing via Personal Access Token, set correct ClientID in opts.
func (a *API) ListWebhooks(opts *ManageWebhooksDefinitionOptions) ([]RegisteredWebhook, error) {
	var resp listWebhooksResponse
	var clientID string
	if opts != nil {
		clientID = opts.ClientID
	}
	err := a.Call("list_webhooks", &listWebhooksRequest{
		OwnerClientID: clientID,
	}, &resp)

	return resp, err
}

// UnregisterWebhook removes webhook with given id from registered webhooks.
//
// When authorizing via Personal Access Token, set correct ClientID in opts.
func (a *API) UnregisterWebhook(id string, opts *ManageWebhooksDefinitionOptions) error {
	var clientID string
	if opts != nil {
		clientID = opts.ClientID
	}
	return a.Call("unregister_webhook", unregisterWebhookRequest{
		ID:            id,
		OwnerClientID: clientID,
	}, &emptyResponse{})
}

// CreateBot allows to create bot and returns its ID.
func (a *API) CreateBot(name string, opts *CreateBotRequestOptions) (string, error) {
	req := createBotRequest{Name: name}
	if opts != nil {
		req.CreateBotRequestOptions = *opts
	}

	if err := validateBotGroupsAssignment(req.Groups); err != nil {
		return "", err
	}
	var resp createBotResponse
	err := a.Call("create_bot", &req, &resp)
	return resp.BotID, err
}

// UpdateBot allows to update bot.
func (a *API) UpdateBot(id string, opts *UpdateBotRequestOptions) error {
	req := updateBotRequest{BotID: id}
	if opts != nil {
		req.UpdateBotRequestOptions = *opts
	}
	if err := validateBotGroupsAssignment(req.Groups); err != nil {
		return err
	}
	return a.Call("update_bot", &req, &emptyResponse{})
}

// DeleteBot deletes bot with given ID.
func (a *API) DeleteBot(id string) error {
	return a.Call("delete_bot", &deleteBotRequest{
		BotID: id,
	}, &emptyResponse{})
}

// ListBots returns list of bots (all or caller's only, depending on getAll parameter).
func (a *API) ListBots(getAll bool, fields []string) ([]*Bot, error) {
	var resp listBotsResponse
	err := a.Call("list_bots", &listBotsRequest{
		All:    getAll,
		Fields: fields,
	}, &resp)

	return resp, err
}

// GetBot returns bot.
func (a *API) GetBot(id string, fields []string) (*Bot, error) {
	var resp getBotResponse
	err := a.Call("get_bot", &getBotRequest{
		BotID:  id,
		Fields: fields,
	}, &resp)

	return resp, err
}

// CreateAgent creates a new Agent with specified parameters within a license.
func (a *API) CreateAgent(id string, fields *AgentFields) (string, error) {
	var resp createAgentResponse
	request := &Agent{
		ID:          id,
		AgentFields: fields,
	}
	err := a.Call("create_agent", request, &resp)

	return resp.ID, err
}

// GetAgent returns the info about an Agent specified by id (i.e. login).
func (a *API) GetAgent(id string, fields []string) (*Agent, error) {
	var resp getAgentResponse
	err := a.Call("get_agent", &getAgentRequest{
		ID:     id,
		Fields: fields,
	}, &resp)

	return resp, err
}

// ListAgents returns all Agents within a license.
func (a *API) ListAgents(groupIDs []int32, fields []string) ([]*Agent, error) {
	var resp listAgentsResponse
	request := &listAgentsRequest{
		Fields: fields,
	}

	if len(groupIDs) > 0 {
		request.Filters = &AgentsFilters{
			GroupIDs: groupIDs,
		}
	}

	err := a.Call("list_agents", request, &resp)
	return resp, err
}

// UpdateAgent updates the properties of an Agent specified by id.
func (a *API) UpdateAgent(id string, fields *AgentFields) error {
	request := &Agent{
		ID:          id,
		AgentFields: fields,
	}
	return a.Call("update_agent", request, &emptyResponse{})
}

// DeleteAgent deletes an Agent specified by id.
func (a *API) DeleteAgent(id string) error {
	return a.Call("delete_agent", &deleteAgentRequest{
		ID: id,
	}, &emptyResponse{})
}

// SuspendAgent suspends an Agent specified by id.
func (a *API) SuspendAgent(id string) error {
	return a.Call("suspend_agent", &suspendAgentRequest{
		ID: id,
	}, &emptyResponse{})
}

// UnsuspendAgent unsuspends an Agent specified by id.
func (a *API) UnsuspendAgent(id string) error {
	return a.Call("unsuspend_agent", &unsuspendAgentRequest{
		ID: id,
	}, &emptyResponse{})
}

// RequestAgentUnsuspension sends a request to license owners and vice owners with an unsuspension request
func (a *API) RequestAgentUnsuspension() error {
	return a.Call("request_agent_unsuspension", nil, &emptyResponse{})
}

// ApproveAgent approves an Agent thus allowing the Agent to use the application.
func (a *API) ApproveAgent(id string) error {
	return a.Call("approve_agent", &approveAgentRequest{
		ID: id,
	}, &emptyResponse{})
}

// RegisterProperty creates private property
func (a *API) RegisterProperty(property *PropertyConfig) error {
	return a.Call("register_property", property, &emptyResponse{})
}

// UnregisterProperty removes private property
func (a *API) UnregisterProperty(name, ownerClientID string) error {
	return a.Call("unregister_property", &unregisterPropertyRequest{
		Name:          name,
		OwnerClientID: ownerClientID,
	}, &emptyResponse{})
}

// PublishProperty publishes private property
func (a *API) PublishProperty(name, ownerClientID string, read, write bool) error {
	accessType := make([]string, 2)
	if read {
		accessType = append(accessType, "read")
	}
	if write {
		accessType = append(accessType, "write")
	}
	return a.Call("publish_property", &publishPropertyRequest{
		Name:          name,
		OwnerClientID: ownerClientID,
		AccessType:    accessType,
	}, &emptyResponse{})
}

// ListProperties return list of properties for given owner_client_id along with their configuration
func (a *API) ListProperties(ownerClientID string) (map[string]*PropertyConfig, error) {
	var resp listPropertiesResponse
	err := a.Call("list_properties", &listPropertiesRequest{
		OwnerClientID: ownerClientID,
	}, &resp)

	return resp, err
}

// CreateGroup creates new group
func (a *API) CreateGroup(name string, agentPriorities map[string]GroupPriority, opts *CreateGroupRequestOptions) (int32, error) {
	req := createGroupRequest{Name: name, AgentPriorities: agentPriorities}
	if opts != nil {
		req.CreateGroupRequestOptions = *opts
	}
	var resp createGroupResponse
	err := a.Call("create_group", &req, &resp)

	return resp.ID, err
}

// UpdateGroup updates existing group
func (a *API) UpdateGroup(id int32, opts *UpdateGroupRequestOptions) error {
	req := updateGroupRequest{ID: id}
	if opts != nil {
		req.UpdateGroupRequestOptions = *opts
	}
	return a.Call("update_group", &req, &emptyResponse{})
}

// DeleteGroup deletes existing group
func (a *API) DeleteGroup(id int32) error {
	return a.Call("delete_group", &deleteGroupRequest{
		ID: id,
	}, &emptyResponse{})
}

// ListGroups lists all existing groups
func (a *API) ListGroups(fields []string) ([]*Group, error) {
	var resp listGroupsResponse
	err := a.Call("list_groups", &listGroupsRequest{
		Fields: fields,
	}, &resp)

	return resp, err
}

// GetGroup returns details about a group specified by its id
func (a *API) GetGroup(id int, fields ...string) (*Group, error) {
	var resp getGroupResponse
	err := a.Call("get_group", &getGroupRequest{
		ID:     id,
		Fields: fields,
	}, &resp)

	return resp, err
}

func validateBotGroupsAssignment(groups []GroupConfig) error {
	for _, group := range groups {
		if group.Priority == DoNotAssign {
			return errors.New("DoNotAssign priority is allowed only as default group priority")
		}
	}

	return nil
}

// ListLicenseProperties returns the properties set within a license.
func (a *API) ListLicenseProperties(opts *ListLicensePropertiesRequestOptions) (Properties, error) {
	req := listLicensePropertiesRequest{}
	if opts != nil {
		req.ListLicensePropertiesRequestOptions = *opts
	}
	var resp Properties
	err := a.Call("list_license_properties", &req, &resp)
	return resp, err
}

// ListWebhookNames returns list of webhooks available in given API version.
func (a *API) ListWebhookNames(version string) ([]*WebhookData, error) {
	var resp []*WebhookData
	err := a.Call("list_webhook_names", &listWebhookNamesRequest{
		Version: version,
	}, &resp)
	return resp, err
}

// EnableLicenseWebhooks enables webhooks for the authorization token's clientID.
//
// When authorizing via Personal Access Token, set correct ClientID in opts.
func (a *API) EnableLicenseWebhooks(opts *ManageWebhooksStateOptions) error {
	var clientID string
	if opts != nil {
		clientID = opts.ClientID
	}
	return a.Call("enable_license_webhooks", &manageWebhooksStateRequest{
		OwnerClientID: clientID,
	}, &emptyResponse{})
}

// DisableLicenseWebhooks disables webhooks for the authorization token's clientID.
//
// When authorizing via Personal Access Token, set correct ClientID in opts.
func (a *API) DisableLicenseWebhooks(opts *ManageWebhooksStateOptions) error {
	var clientID string
	if opts != nil {
		clientID = opts.ClientID
	}
	return a.Call("disable_license_webhooks", &manageWebhooksStateRequest{
		OwnerClientID: clientID,
	}, &emptyResponse{})
}

// GetLicenseWebhooksState retrieves webhooks' state for the authorization token's clientID.
//
// When authorizing via Personal Access Token, set correct ClientID in opts.
func (a *API) GetLicenseWebhooksState(opts *ManageWebhooksStateOptions) (*WebhooksState, error) {
	var clientID string
	if opts != nil {
		clientID = opts.ClientID
	}
	var resp *WebhooksState
	err := a.Call("get_license_webhooks_state", &manageWebhooksStateRequest{
		OwnerClientID: clientID,
	}, &resp)
	return resp, err
}

// UpdateLicenseProperties updates the properties set within a license.
func (a *API) UpdateLicenseProperties(props Properties) error {
	return a.Call("update_license_properties", &updateLicensePropertiesRequest{
		Properties: props,
	}, &emptyResponse{})
}

// UpdateGroupProperties updates the properties set within a group.
func (a *API) UpdateGroupProperties(id int, props Properties) error {
	return a.Call("update_group_properties", &updateGroupPropertiesRequest{
		ID:         id,
		Properties: props,
	}, &emptyResponse{})
}

// DeleteLicenseProperties deletes the properties set within a license.
func (a *API) DeleteLicenseProperties(props map[string][]string) error {
	return a.Call("delete_license_properties", &deleteLicensePropertiesRequest{
		Properties: props,
	}, &emptyResponse{})
}

// DeleteGroupProperties deletes the properties set within a group.
func (a *API) DeleteGroupProperties(id int, props map[string][]string) error {
	return a.Call("delete_group_properties", &deleteGroupPropertiesRequest{
		ID:         id,
		Properties: props,
	}, &emptyResponse{})
}

// AddAutoAccess creates an auto access data structure.
func (a *API) AddAutoAccess(access Access, conditions AutoAccessConditions, opts *AddAutoAccessRequestOptions) (string, error) {
	req := addAutoAccessRequest{Access: access, Conditions: conditions}
	if opts != nil {
		req.AddAutoAccessRequestOptions = *opts
	}
	var resp addAutoAccessResponse
	err := a.Call("add_auto_access", &req, &resp)
	return resp.ID, err
}

// UpdateAutoAccess updates an existing auto access.
func (a *API) UpdateAutoAccess(id string, opts *UpdateAutoAccessRequestOptions) error {
	req := updateAutoAccessRequest{ID: id}
	if opts != nil {
		req.UpdateAutoAccessRequestOptions = *opts
	}
	return a.Call("update_auto_access", &req, &emptyResponse{})
}

// DeleteAutoAccess deletes an existing auto access.
func (a *API) DeleteAutoAccess(id string) error {
	return a.Call("delete_auto_access", &deleteAutoAccessRequest{ID: id}, &emptyResponse{})
}

// ListAutoAccesses returns all existing auto access.
func (a *API) ListAutoAccesses() ([]*AutoAccess, error) {
	var resp []*AutoAccess
	err := a.Call("list_auto_accesses", &listAutoAccessesRequest{}, &resp)
	return resp, err
}

// CheckProductLimitsForPlan compares your organization's current resources with a given plan and returns those which exceeded the called plan's limits.
func (a *API) CheckProductLimitsForPlan(plan string) (PlanLimits, error) {
	var resp PlanLimits
	err := a.Call("check_product_limits_for_plan", &checkProductLimitsForPlanRequest{
		Plan: plan,
	}, &resp)
	return resp, err
}

// ListChannels returns the summary of communication channels for your LiveChat product.
func (a *API) ListChannels() (ChannelActivity, error) {
	var resp ChannelActivity
	err := a.Call("list_channels", &listChannelsRequest{}, &resp)
	return resp, err
}

// CreateTag creates a new tag
func (a *API) CreateTag(name string, groupIDs []int) error {
	return a.Call("create_tag", &createTagRequest{
		Name:     name,
		GroupIDs: groupIDs,
	}, &emptyResponse{})
}

// DeleteTag deletes an existing tag
func (a *API) DeleteTag(name string) error {
	return a.Call("delete_tag", &deleteTagRequest{
		Name: name,
	}, &emptyResponse{})
}

// ListTags returns tags assigned to requested groups
func (a *API) ListTags(groupIDs []int) ([]*Tag, error) {
	var resp []*Tag
	err := a.Call("list_tags", &listTagsRequest{
		GroupIDs: groupIDs,
	}, &resp)
	return resp, err
}

// UpdateTag updates an existing tag
func (a *API) UpdateTag(name string, groupIDs []int) error {
	return a.Call("update_tag", &updateTagRequest{
		Name:     name,
		GroupIDs: groupIDs,
	}, &emptyResponse{})
}

// Lists properties of groups
func (a *API) ListGroupsProperties(groupIDs []int, opts *ListGroupsPropertiesRequestOptions) ([]GroupProperties, error) {
	req := listGroupsPropertiesRequest{GroupIDs: groupIDs}
	if opts != nil {
		req.ListGroupsPropertiesRequestOptions = *opts
	}
	var resp []GroupProperties
	err := a.Call("list_groups_properties", &req, &resp)
	return resp, err
}

// Reactivates bounced email
func (a *API) ReactivateEmail(agentID string) error {
	return a.Call("reactivate_email", &reactivateEmailRequest{
		AgentID: agentID,
	}, &emptyResponse{})
}

// Updates company details
func (a *API) UpdateCompanyDetails(companyDetails CompanyDetails, enrich bool) error {
	return a.Call("update_company_details", &updateCompanyDetailsRequest{
		CompanyDetails: companyDetails,
		Enrich:         enrich,
	}, &emptyResponse{})
}
