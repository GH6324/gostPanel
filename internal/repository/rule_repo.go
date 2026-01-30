	})
}

// StopByTunnelIDs 停止指定隧道列表关联的所有规则
func (r *RuleRepository) StopByTunnelIDs(tunnelIDs []uint) error {
	if len(tunnelIDs) == 0 {
		return nil
	}
	return r.DB.Model(&model.GostRule{}).
		Where("tunnel_id IN ? AND status = ?", tunnelIDs, model.RuleStatusRunning).
		Update("status", model.RuleStatusStopped).Error
}
