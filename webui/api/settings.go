package api

import (
	"encoding/json"
	"net/http"

	"github.com/alcaprophet/fwalizer/config"
)

func (d *Deps) handleGetSettings(w http.ResponseWriter, r *http.Request) {
	settings, err := d.Store.GetSettings()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, settings)
}

func (d *Deps) handlePutSettings(w http.ResponseWriter, r *http.Request) {
	var settings map[string]string
	if err := json.NewDecoder(r.Body).Decode(&settings); err != nil {
		writeError(w, http.StatusBadRequest, "请求格式错误")
		return
	}
	for k, v := range settings {
		if err := d.Store.SetSetting(k, v); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}
	d.notifyReload()
	writeJSON(w, http.StatusOK, map[string]string{"message": "保存成功"})
}

// configExport 配置导出结构（凭据不导出）
type configExport struct {
	Version  int                   `json:"version"`
	Targets  []config.TargetConfig `json:"targets"`
	Rules    []config.DomainRule   `json:"rules"`
	Settings map[string]string     `json:"settings"`
}

func (d *Deps) handleConfigExport(w http.ResponseWriter, r *http.Request) {
	targets, err := d.Store.GetTargets()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	rules, err := d.Store.GetRules()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	settings, err := d.Store.GetSettings()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// 凭据字段不导出（安全考虑）
	delete(settings, "tc_access_id")
	delete(settings, "tc_access_key")
	delete(settings, "ali_access_id")
	delete(settings, "ali_access_key")

	export := configExport{
		Version:  1,
		Targets:  targets,
		Rules:    rules,
		Settings: settings,
	}

	w.Header().Set("Content-Disposition", "attachment; filename=fwalizer-config.json")
	writeJSON(w, http.StatusOK, export)
}

// configImport 配置导入结构
type configImport struct {
	Version  int                   `json:"version"`
	Targets  []config.TargetConfig `json:"targets"`
	Rules    []config.DomainRule   `json:"rules"`
	Settings map[string]string     `json:"settings"`
}

func (d *Deps) handleConfigImport(w http.ResponseWriter, r *http.Request) {
	var imp configImport
	if err := json.NewDecoder(r.Body).Decode(&imp); err != nil {
		writeError(w, http.StatusBadRequest, "JSON 格式错误")
		return
	}
	if imp.Version != 1 {
		writeError(w, http.StatusBadRequest, "不支持的配置版本")
		return
	}

	// 清空旧数据并写入新数据
	if err := d.Store.ClearAll(); err != nil {
		writeError(w, http.StatusInternalServerError, "清空旧配置失败: "+err.Error())
		return
	}
	if err := d.Store.BatchAddTargets(imp.Targets); err != nil {
		writeError(w, http.StatusInternalServerError, "导入目标失败: "+err.Error())
		return
	}
	if err := d.Store.BatchAddRules(imp.Rules); err != nil {
		writeError(w, http.StatusInternalServerError, "导入规则失败: "+err.Error())
		return
	}
	for k, v := range imp.Settings {
		// 跳过凭据字段（不导入）
		if k == "tc_access_id" || k == "tc_access_key" || k == "ali_access_id" || k == "ali_access_key" {
			continue
		}
		if err := d.Store.SetSetting(k, v); err != nil {
			writeError(w, http.StatusInternalServerError, "导入设置失败: "+err.Error())
			return
		}
	}

	d.notifyReload()
	writeJSON(w, http.StatusOK, map[string]string{"message": "导入成功"})
}
