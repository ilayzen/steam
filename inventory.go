package steam

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
)

const (
	InventoryEndpoint           = "http://steamcommunity.com/inventory/%d/%d/%d?"
	contextInventoryEndpoint    = "profiles/%s/inventory/"
	steamTimeAPI                = "https://api.steampowered.com/ITwoFactorService/QueryTime/v0001"
	getConfirmationListEndpoint = SteamcommunityURL + "mobileconf/getlist?p=%s&a=%s&k=%s&t=%s&m=%s&tag=%s"
	acceptConfirmation          = SteamcommunityURL + "mobileconf/ajaxop?op=%s&p=%s&a=%s&k=%s&t=%s&m=react&tag=%s&cid=%s&ck=%s"
	conf                        = "conf"
)

type ItemTag struct {
	Category              string `json:"category"`
	InternalName          string `json:"internal_name"`
	LocalizedCategoryName string `json:"localized_category_name"`
	LocalizedTagName      string `json:"localized_tag_name"`
}

// Due to the JSON being string, etc... we cannot re-use EconItem
// Also, "assetid" is included as "id" not as assetid.
type InventoryItem struct {
	AppID      uint32        `json:"appid"`
	ContextID  uint64        `json:"contextid"`
	AssetID    uint64        `json:"id,string,omitempty"`
	ClassID    uint64        `json:"classid,string,omitempty"`
	InstanceID uint64        `json:"instanceid,string,omitempty"`
	Amount     uint64        `json:"amount,string"`
	Desc       *EconItemDesc `json:"-"` /* May be nil  */
}

type InventoryContext struct {
	ID         uint64 `json:"id,string"` /* Apparently context id needs at least 64 bits...  */
	AssetCount uint32 `json:"asset_count"`
	Name       string `json:"name"`
}

type InventoryAppStats struct {
	AppID            uint64                       `json:"appid"`
	Name             string                       `json:"name"`
	AssetCount       uint32                       `json:"asset_count"`
	Icon             string                       `json:"icon"`
	Link             string                       `json:"link"`
	InventoryLogo    string                       `json:"inventory_logo"`
	TradePermissions string                       `json:"trade_permissions"`
	Contexts         map[string]*InventoryContext `json:"rgContexts"`
}

var inventoryContextRegexp = regexp.MustCompile("var g_rgAppContextData = (.*?);")

func (session *Session) fetchInventory(
	sid SteamID,
	appID, contextID, startAssetID uint64,
	filters []Filter,
	items *[]InventoryItem,
) (hasMore bool, lastAssetID uint64, err error) {
	params := url.Values{
		"l": {session.language},
	}

	if startAssetID != 0 {
		params.Set("start_assetid", strconv.FormatUint(startAssetID, 10))
		params.Set("count", "75")
	} else {
		params.Set("count", "250")
	}

	resp, err := session.client.Get(fmt.Sprintf(InventoryEndpoint, sid, appID, contextID) + params.Encode())
	if resp != nil {
		defer resp.Body.Close()
	}

	if err != nil {
		return false, 0, err
	}

	type Asset struct {
		AppID      uint32 `json:"appid"`
		ContextID  uint64 `json:"contextid,string"`
		AssetID    uint64 `json:"assetid,string"`
		ClassID    uint64 `json:"classid,string"`
		InstanceID uint64 `json:"instanceid,string"`
		Amount     uint64 `json:"amount,string"`
	}

	type Response struct {
		Assets              []Asset         `json:"assets"`
		Descriptions        []*EconItemDesc `json:"descriptions"`
		Success             int             `json:"success"`
		HasMore             int             `json:"more_items"`
		LastAssetID         string          `json:"last_assetid"`
		TotalInventoryCount int             `json:"total_inventory_count"`
		ErrorMsg            string          `json:"error"`
	}

	var response Response
	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return false, 0, err
	}

	if response.Success == 0 {
		if len(response.ErrorMsg) != 0 {
			return false, 0, errors.New(response.ErrorMsg)
		}

		return false, 0, nil // empty inventory
	}

	// Fill in descriptions map, where key
	// is "<CLASS_ID>_<INSTANCE_ID>" pattern, and
	// value is position on asset description in
	// response.Descriptions array
	//
	// We need it for fast asset's description
	// searching in future
	descriptions := make(map[string]int)
	for i, desc := range response.Descriptions {
		key := fmt.Sprintf("%d_%d", desc.ClassID, desc.InstanceID)
		descriptions[key] = i
	}

	for _, asset := range response.Assets {
		var desc *EconItemDesc

		key := fmt.Sprintf("%d_%d", asset.ClassID, asset.InstanceID)
		if d, ok := descriptions[key]; ok {
			desc = response.Descriptions[d]
		}

		item := InventoryItem{
			AppID:      asset.AppID,
			ContextID:  asset.ContextID,
			AssetID:    asset.AssetID,
			ClassID:    asset.ClassID,
			InstanceID: asset.InstanceID,
			Amount:     asset.Amount,
			Desc:       desc,
		}

		add := true
		for _, filter := range filters {
			add = filter(&item)
			if !add {
				break
			}
		}

		if add {
			*items = append(*items, item)
		}
	}

	hasMore = response.HasMore != 0
	if !hasMore {
		return hasMore, 0, nil
	}

	lastAssetID, err = strconv.ParseUint(response.LastAssetID, 10, 64)
	if err != nil {
		return hasMore, 0, err
	}

	return hasMore, lastAssetID, nil
}

func (session *Session) GetInventory(sid SteamID, appID, contextID uint64) ([]InventoryItem, error) {
	filters := []Filter{}

	return session.GetFilterableInventory(sid, appID, contextID, filters)
}

func (session *Session) GetFilterableInventory(sid SteamID, appID, contextID uint64, filters []Filter) ([]InventoryItem, error) {
	items := []InventoryItem{}
	startAssetID := uint64(0)

	for {
		hasMore, lastAssetID, err := session.fetchInventory(sid, appID, contextID, startAssetID, filters, &items)
		if err != nil {
			return nil, err
		}

		if !hasMore {
			break
		}

		startAssetID = lastAssetID
	}

	return items, nil
}

func (session *Session) GetInventoryAppStats(sid SteamID) (map[string]InventoryAppStats, error) {
	resp, err := session.client.Get("https://steamcommunity.com/profiles/" + sid.ToString() + "/inventory")
	if resp != nil {
		defer resp.Body.Close()
	}

	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	m := inventoryContextRegexp.FindSubmatch(body)
	if m == nil || len(m) != 2 {
		return nil, err
	}

	inven := map[string]InventoryAppStats{}
	if err = json.Unmarshal(m[1], &inven); err != nil {
		return nil, err
	}

	return inven, nil
}

func (session *Session) GetInventoryContext(steamID string) (*SteamInventoryContext, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf(SteamcommunityURL+contextInventoryEndpoint, steamID), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create a request, err: %v", err)
	}

	resp, err := session.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get html page, err: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("cannot read html page, err: %v", err)
	}

	re := regexp.MustCompile(`g_rgAppContextData\s*=\s*(\{.*?\});`)
	match := re.FindStringSubmatch(string(body))

	if len(match) == 0 {
		return nil, fmt.Errorf("inventory context is empty")
	}

	if len(match) < 2 {
		return nil, fmt.Errorf("cannot get g_rgAppContextData in html page")
	}

	var invContext SteamInventoryContext
	if err := json.Unmarshal([]byte(match[1]), &invContext); err != nil {
		return nil, fmt.Errorf("failed to unmarshal json, err %v", err)
	}

	return &invContext, nil
}

func generateConfirmationHashForTime(identitySecret string, tag string, timestamp int64) (string, error) {
	decodedSecret, err := base64.StdEncoding.DecodeString(identitySecret)
	if err != nil {
		return "", fmt.Errorf("failed to decode identitySecret: %v", err)
	}

	if len(tag) > 32 {
		tag = tag[:32]
	}

	n2 := 8 + len(tag)
	data := make([]byte, n2)
	for i := 7; i >= 0; i-- {
		data[i] = byte(timestamp & 0xFF)
		timestamp >>= 8

	}
	if tag != "" {
		copy(data[8:], []byte(tag))
	}

	h := hmac.New(sha1.New, decodedSecret)
	h.Write(data)
	hashedData := h.Sum(nil)

	encodedData := base64.StdEncoding.EncodeToString(hashedData)
	return url.QueryEscape(encodedData), nil
}

func (s *Session) FetchConfirmations(identitySecret string) (*ConfirmationResponse, error) {
	timestamp, err := s.getSteamTime()
	if err != nil {
		return nil, fmt.Errorf("failed to get Steam time: %w", err)
	}

	hash, err := generateConfirmationHashForTime(identitySecret, conf, timestamp)
	if err != nil {
		return nil, fmt.Errorf("failed to generate confirmation hash: %w", err)
	}

	steamID := s.GetSteamID()

	confListEndpoint := fmt.Sprintf(getConfirmationListEndpoint, s.deviceID, steamID.ToString(), hash, strconv.FormatInt(timestamp, 10), "react", conf)

	req, err := http.NewRequest(http.MethodGet, confListEndpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	confirmations := ConfirmationResponse{}
	if err := json.Unmarshal(body, &confirmations); err != nil {
		return nil, fmt.Errorf("error parsing response JSON: %w", err)
	}

	return &confirmations, nil
}

func (s *Session) getSteamTime() (int64, error) {
	req, err := http.NewRequest(http.MethodPost, steamTimeAPI, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("Steam API returned status code %d", resp.StatusCode)
	}

	var result SteamTimeResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("failed to parse Steam time response: %v", err)
	}

	return result.SteamTime.ServerTime, nil
}

func (s *Session) AcceptConfirmation(identitySecret string) (*ConfirmationResponse, error) {
	timestamp, err := s.getSteamTime()
	if err != nil {
		return nil, fmt.Errorf("failed to get Steam time: %w", err)
	}

	hash, err := generateConfirmationHashForTime(identitySecret, conf, timestamp)
	if err != nil {
		return nil, fmt.Errorf("failed to generate confirmation hash: %w", err)
	}

	steamID := s.GetSteamID()

	confListEndpoint := fmt.Sprintf(getConfirmationListEndpoint, s.deviceID, steamID.ToString(), hash, strconv.FormatInt(timestamp, 10), "react", conf)

	req, err := http.NewRequest(http.MethodGet, confListEndpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	confirmations := ConfirmationResponse{}
	if err := json.Unmarshal(body, &confirmations); err != nil {
		return nil, fmt.Errorf("error parsing response JSON: %w", err)
	}

	return &confirmations, nil
}

func (s *Session) SendConfirmationAjax(conf *Confirmation, tag, is string) (*ConfirmationAcceptResponse, error) {
	//tag can be only reject or accept
	op := "cancel"
	if tag == "accept" {
		op = "allow"
	}

	timestamp, err := s.getSteamTime()
	if err != nil {
		return nil, fmt.Errorf("failed to get Steam time: %w", err)
	}

	hash, err := generateConfirmationHashForTime(is, tag, timestamp)
	if err != nil {
		return nil, fmt.Errorf("failed to generate confirmation hash: %w", err)
	}

	steamID := s.GetSteamID()

	confListEndpoint := fmt.Sprintf(acceptConfirmation, op, s.deviceID, steamID.ToString(), hash, strconv.FormatInt(timestamp, 10), tag, conf.ID, conf.Nonce)

	req, err := http.NewRequest(http.MethodGet, confListEndpoint, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	confAccessResponse := &ConfirmationAcceptResponse{}
	err = json.Unmarshal(body, confAccessResponse)
	if err != nil {
		return nil, err
	}

	return confAccessResponse, nil
}
