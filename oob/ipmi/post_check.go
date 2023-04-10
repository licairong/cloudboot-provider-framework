package ipmi

import (
	"fmt"
	"github.com/licairong/cloudboot-provider-framework/oob"
	"github.com/licairong/cloudboot-provider-framework/util"
	"strings"
)

// checkNetwork 检查实际的OOB网络是否与预期的配置相符
func (w *worker) checkNetwork(sett *oob.NetworkSetting) (items []*hardware.CheckingItem) {
	if sett == nil || sett.IPSrc == "" {
		return nil
	}
	network, err := w.Network()
	if err != nil {
		return []*util.CheckingItem{
			{
				Title:   "Network",
				Matched: util.MatchedUnknown,
				Error:   err.Error(),
			},
		}
	}

	items = append(items,
		util.NewCheckingHelper("IP Source", sett.IPSrc, strings.ToLower(network.IPSrc)).Matcher(util.ContainsMatch).Do(),
	)

	if sett.IPSrc == oob.Static {
		items = append(items,
			util.NewCheckingHelper("IP", sett.StaticIP.IP, network.IP).Do(),
			util.NewCheckingHelper("Netmask", sett.StaticIP.Netmask, network.Netmask).Do(),
			util.NewCheckingHelper("Gateway", sett.StaticIP.Gateway, network.Gateway).Do(),
		)
	}
	return items
}

// checkUser 检查实际的OOB用户是否与预期配置相符
func (w *worker) checkUser(sett *oob.UserSetting) (items []*util.CheckingItem) {
	if sett == nil {
		return nil
	}
	users, err := w.Users()
	if err != nil {
		return []*util.CheckingItem{
			{
				Title:   "Users",
				Matched: util.MatchedUnknown,
				Error:   err.Error(),
			},
		}
	}

	for _, settUser := range []*oob.UserSettingItem(*sett) {
		item := util.CheckingItem{
			Title:    "Create User",
			Expected: fmt.Sprintf("%s@%s", settUser.Username, oob.StringUserLevel(settUser.PrivilegeLevel)),
			Matched:  util.MatchedNO,
		}
		idx := w.findUserIndexByName(users, settUser.Username)
		// 检查目标用户是否存在
		if idx < 0 {
			item.Actual = "Missing"
			items = append(items, &item)
			continue
		}
		// 检查目标用户权限级别
		if users[idx].Access == nil || users[idx].Access.PrivilegeLevel < 0 {
			item.Actual = fmt.Sprintf("%s@unknown", settUser.Username)
			items = append(items, &item)
			continue
		}

		item.Actual = fmt.Sprintf("%s@%s", users[idx].Name, oob.StringUserLevel(users[idx].Access.PrivilegeLevel))
		if item.Actual == item.Expected {
			item.Matched = util.MatchedYES
		}
		items = append(items, &item)
	}
	return items
}
