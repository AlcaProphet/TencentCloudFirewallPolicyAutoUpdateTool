package provider

import "sync"

// credentials 凭据持有者（由 app 层在创建 Provider 前设置）
var (
	credMu      sync.RWMutex
	tcAccessID  string
	tcAccessKey string
	aliAccessID  string
	aliAccessKey string
)

// SetCredentials 设置云厂商凭据（由 app 层调用）
func SetCredentials(tcID, tcKey, aliID, aliKey string) {
	credMu.Lock()
	defer credMu.Unlock()
	tcAccessID = tcID
	tcAccessKey = tcKey
	aliAccessID = aliID
	aliAccessKey = aliKey
}

func getTCAccessID() string {
	credMu.RLock()
	defer credMu.RUnlock()
	return tcAccessID
}

func getTCAccessKey() string {
	credMu.RLock()
	defer credMu.RUnlock()
	return tcAccessKey
}

func getAliAccessID() string {
	credMu.RLock()
	defer credMu.RUnlock()
	return aliAccessID
}

func getAliAccessKey() string {
	credMu.RLock()
	defer credMu.RUnlock()
	return aliAccessKey
}
