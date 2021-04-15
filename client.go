package op

import (
	"fmt"
	"os"

	"github.com/pkg/errors"

	"github.com/go-resty/resty/v2"
)

const (
	TokenKEY = "OPASSWORD_ACCESS_TOKEN"
	VERSION  = "/v1"
)

// ItemCategory Represents the template of the Item
type ItemCategory string

const (
	Password   ItemCategory = "PASSWORD"
	SecureNote ItemCategory = "SECURE_NOTE"
)

type Client struct {
	conn     *resty.Client
	endpoint string
	token    string

	Vault *VaultClient
}

type ItemClient struct {
	client *Client

	Note     *NoteClient
	Password *PasswordClient
}

type VaultClient struct {
	client      *Client
	currentID   string
	currentName string

	Item *ItemClient
}

type ClientOption func(*Client)

func WithTokenFromEnv() ClientOption {
	return func(c *Client) {
		value := os.Getenv(TokenKEY)
		c.token = value
	}
}

func WithToken(value string) ClientOption {
	return func(c *Client) {
		c.token = value
	}
}

func NewClient(endpoint string, options ...ClientOption) (*Client, error) {
	client := resty.New()
	c := &Client{
		conn:     client,
		endpoint: endpoint + VERSION,
	}
	note := &NoteClient{client: c}
	password := &PasswordClient{client: c}
	item := &ItemClient{
		client:   c,
		Note:     note,
		Password: password,
	}
	vault := &VaultClient{
		client: c,
		Item:   item,
	}
	c.Vault = vault

	for _, option := range options {
		option(c)
	}
	c.conn.SetAuthToken(c.token).SetHeader("Accept", "application/json")
	return c, nil
}

func (c *VaultClient) Get(name string) (*VaultClient, error) {
	var vaults []Vault
	resp, err := c.client.conn.R().SetResult(&vaults).Get(c.client.endpoint + "/vaults")
	if err != nil {
		return nil, errors.Wrapf(err, "connection error")
	}
	if resp.IsError() {
		return nil, errors.Wrapf(err, "http error")
	}
	for _, v := range vaults {
		if name == v.Name() {
			c.currentName = name
			c.currentID = v.ID()
			return c, nil
		}
	}
	return nil, errors.New("no results")
}

func (c *VaultClient) UUID() string {
	return c.currentID
}

func (c *VaultClient) Name() string {
	return c.currentName
}

type Vault map[string]interface{}

func (v Vault) Name() string {
	name, _ := v["name"].(string)
	return name
}

func (v Vault) ID() string {
	id, _ := v["id"].(string)
	return id
}

func (c *VaultClient) SetID(uuid string) {
	c.currentID = uuid
}

type Item map[string]interface{}

func (i Item) Tags() []string {
	// FIXME: Convert to []string directly if possible
	t, ok := i["tags"].([]interface{})
	if !ok {
		return []string{}
	}
	tags := make([]string, len(t))

	for n, tag := range t {
		tags[n] = tag.(string)
	}
	return tags
}

func (i Item) Title() string {
	title, _ := i["title"].(string)
	return title
}

func (i Item) Category() ItemCategory {
	category, _ := i["category"].(ItemCategory)
	return category
}

func (i Item) ID() string {
	id, _ := i["id"].(string)
	return id
}

type Filter func([]Item) []Item

func stringExists(slice []string, val string) bool {
	for _, str := range slice {
		if str == val {
			return true
		}
	}
	return false
}

func itemContainsTag(item Item, tag string) bool {
	return stringExists(item.Tags(), tag)
}

func FilterByTags(tags ...string) Filter {
	return func(items []Item) []Item {
		newItems := make([]Item, 0)
	next:
		for _, i := range items {
			for _, tag := range tags {
				if !itemContainsTag(i, tag) {
					continue next
				}
			}
			newItems = append(newItems, i)
		}
		return newItems
	}
}

func FilterByTitle(name string) Filter {
	return func(items []Item) []Item {
		newItems := make([]Item, 0)
		for _, i := range items {
			if i.Title() == name {
				newItems = append(newItems, i)
			}
		}
		return newItems
	}
}

func FilterByCategory(name ItemCategory) Filter {
	return func(items []Item) []Item {
		newItems := make([]Item, 0)
		for _, i := range items {
			if i.Category() == name {
				newItems = append(newItems, i)
			}
		}
		return newItems
	}
}

func (c *ItemClient) GetDetails(uuid string) (Item, error) {
	var item Item
	vaultID := c.client.Vault.UUID()
	path := fmt.Sprintf("/vaults/%s/items/%s", vaultID, uuid)
	resp, err := c.client.conn.R().SetResult(&item).Get(c.client.endpoint + path)
	if err != nil {
		return nil, errors.Wrapf(err, "connection error")
	}
	if resp.IsError() {
		return nil, errors.Wrapf(err, "http error")
	}
	return item, nil
}

func (c *ItemClient) Get(filters ...Filter) ([]Item, error) {
	var items []Item
	uuid := c.client.Vault.UUID()
	path := fmt.Sprintf("/vaults/%s/items", uuid)
	resp, err := c.client.conn.R().SetResult(&items).Get(c.client.endpoint + path)
	if err != nil {
		return nil, errors.Wrapf(err, "connection error")
	}
	if resp.IsError() {
		return nil, errors.Wrapf(err, "http error")
	}
	itemsDetailed := make([]Item, len(items))
	for n, i := range items {
		item, err := c.GetDetails(i.ID())
		if err != nil {
			return nil, err
		}
		itemsDetailed[n] = item
	}
	filtered := applyFilters(itemsDetailed, filters...)
	return filtered, nil
}

type NoteClient struct {
	client *Client
}

type PasswordClient struct {
	client *Client
}

type ItemCreateOption func(Item) Item

func WithTags(tags ...string) ItemCreateOption {
	return func(item Item) Item {
		return item.AppendTags(tags...)
	}
}

func WithFields(fields ...map[string]interface{}) ItemCreateOption {
	return func(item Item) Item {
		item = item.AppendFields(fields...)
		return item
	}
}

func WithCategory(category ItemCategory) ItemCreateOption {
	return func(item Item) Item {
		item = item.SetCategory(category)
		return item
	}
}

func WithVaultID(uuid string) ItemCreateOption {
	return func(item Item) Item {
		item = item.SetVaultID(uuid)
		return item
	}
}

func (i Item) AppendTags(tags ...string) Item {
	t := i.Tags()
	t = append(t, tags...)
	i["tags"] = t
	return i
}

func (i Item) Fields() []map[string]interface{} {
	fields, ok := i["fields"].([]map[string]interface{})
	if !ok {
		return make([]map[string]interface{}, 0)
	}
	return fields
}

func (i Item) SetVaultID(uuid string) Item {
	i["vault"] = map[string]string{
		"id": uuid,
	}
	return i
}

func (i Item) SetCategory(category ItemCategory) Item {
	i["category"] = category
	return i
}

func (i Item) AppendFields(fields ...map[string]interface{}) Item {
	f := i.Fields()
	f = append(f, fields...)
	i["fields"] = f
	return i
}

func NewItem(title string, options ...ItemCreateOption) Item {
	item := Item{
		"title":  title,
		"tags":   make([]string, 0),
		"fields": make([]map[string]interface{}, 0),
	}
	for _, option := range options {
		item = option(item)
	}
	return item
}

func (c *NoteClient) New(title, content string, options ...ItemCreateOption) Item {
	uuid := c.client.Vault.UUID()
	item := NewItem(
		title,
		WithVaultID(uuid),
		WithCategory(SecureNote),
		WithFields(
			map[string]interface{}{
				"id":      "notesPlain",
				"label":   "notesPlain",
				"purpose": "NOTES",
				"type":    "STRING",
				"value":   content,
			},
		))
	for _, option := range options {
		item = option(item)
	}
	return item
}

func (c *NoteClient) Create(title, content string, options ...ItemCreateOption) (*Item, error) {
	note := c.New(title, content, options...)
	return c.client.Vault.Item.Add(note)
}

func (c *PasswordClient) New(title, secret string, options ...ItemCreateOption) Item {
	uuid := c.client.Vault.UUID()
	item := NewItem(
		title,
		WithVaultID(uuid),
		WithCategory(Password),
		WithFields(
			map[string]interface{}{
				"id":      "password",
				"label":   "password",
				"purpose": "PASSWORD",
				"type":    "CONCEALED",
				"value":   secret,
			},
		))
	for _, option := range options {
		item = option(item)
	}
	return item
}

func (c *PasswordClient) Create(title, secret string, options ...ItemCreateOption) (*Item, error) {
	note := c.New(title, secret, options...)
	return c.client.Vault.Item.Add(note)
}

func (c *ItemClient) Add(item Item) (*Item, error) {
	uuid := c.client.Vault.UUID()
	path := fmt.Sprintf("/vaults/%s/items", uuid)
	endpoint := c.client.endpoint + path
	result := new(Item)
	resp, err := c.client.conn.R().SetResult(result).SetBody(item).Post(endpoint)
	if err != nil {
		return nil, errors.Wrapf(err, "connection error")
	}
	if resp.IsError() {
		return nil, errors.Wrapf(err, "http error")
	}
	return result, nil
}

func applyFilters(items []Item, filters ...Filter) []Item {
	if len(filters) == 0 {
		return items
	}
	filter := filters[0]
	if len(filters) == 1 {
		return filter(items)
	}
	items = filter(items)
	rest := filters[1:]
	return applyFilters(items, rest...)
}
