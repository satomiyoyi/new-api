package controller

import (
	"net/http"
	"runtime"
	"strconv"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/middleware"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
)

// GetMonitorOverview 获取监控概览
func GetMonitorOverview(c *gin.Context) {
	overview, err := model.GetMonitorOverview()
	if err != nil {
		common.ApiErrorMsg(c, "获取监控概览失败: "+err.Error())
		return
	}

	// 活跃连接数
	stats := middleware.GetStats()
	overview.ActiveConnections = stats.ActiveConnections

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    overview,
	})
}

// GetMonitorSystem 获取系统资源信息
func GetMonitorSystem(c *gin.Context) {
	systemStatus := common.GetSystemStatus()
	diskInfo := common.GetDiskSpaceInfo()
	cacheStats := common.GetDiskCacheStats()

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	hitRate := float64(0)
	totalCacheOps := cacheStats.DiskCacheHits + cacheStats.MemoryCacheHits
	if totalCacheOps > 0 {
		hitRate = float64(totalCacheOps) / float64(totalCacheOps+cacheStats.ActiveDiskFiles) * 100
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"cpu_usage":      systemStatus.CPUUsage,
			"memory_usage":   systemStatus.MemoryUsage,
			"disk_usage":     systemStatus.DiskUsage,
			"disk_total":     diskInfo.Total,
			"disk_used":      diskInfo.Used,
			"disk_free":      diskInfo.Free,
			"num_goroutine":  runtime.NumGoroutine(),
			"go_alloc":       memStats.Alloc,
			"go_sys":         memStats.Sys,
			"go_total_alloc": memStats.TotalAlloc,
			"go_num_gc":      memStats.NumGC,
			"cache_hit_rate": hitRate,
		},
	})
}

// GetMonitorChannels 获取渠道健康数据
func GetMonitorChannels(c *gin.Context) {
	minutes := parseMinutes(c, 60)
	sinceTimestamp := time.Now().Add(-time.Duration(minutes) * time.Minute).Unix()

	data, err := model.GetChannelHealth(sinceTimestamp)
	if err != nil {
		common.ApiErrorMsg(c, "获取渠道健康数据失败: "+err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    data,
	})
}

// GetMonitorModels 获取模型性能数据
func GetMonitorModels(c *gin.Context) {
	minutes := parseMinutes(c, 60)
	sinceTimestamp := time.Now().Add(-time.Duration(minutes) * time.Minute).Unix()

	data, err := model.GetModelPerformance(sinceTimestamp)
	if err != nil {
		common.ApiErrorMsg(c, "获取模型性能数据失败: "+err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    data,
	})
}

// GetMonitorTrends 获取趋势数据
func GetMonitorTrends(c *gin.Context) {
	minutes := parseMinutes(c, 60)
	granularity := int64(60)
	if g, err := strconv.ParseInt(c.DefaultQuery("granularity", "60"), 10, 64); err == nil && g > 0 {
		granularity = g
	}
	channelId := 0
	if ch, err := strconv.Atoi(c.DefaultQuery("channel_id", "0")); err == nil {
		channelId = ch
	}
	modelName := c.DefaultQuery("model_name", "")
	sinceTimestamp := time.Now().Add(-time.Duration(minutes) * time.Minute).Unix()

	data, err := model.GetMonitorTrends(model.TrendFilter{
		SinceTimestamp: sinceTimestamp,
		Granularity:    granularity,
		ChannelId:      channelId,
		ModelName:      modelName,
	})
	if err != nil {
		common.ApiErrorMsg(c, "获取趋势数据失败: "+err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    data,
	})
}

// GetMonitorLatency 获取延迟分布
func GetMonitorLatency(c *gin.Context) {
	minutes := parseMinutes(c, 60)
	sinceTimestamp := time.Now().Add(-time.Duration(minutes) * time.Minute).Unix()

	data, err := model.GetMonitorLatency(sinceTimestamp)
	if err != nil {
		common.ApiErrorMsg(c, "获取延迟分布失败: "+err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    data,
	})
}

// GetMonitorTopUsers 获取 Top N 用户
func GetMonitorTopUsers(c *gin.Context) {
	minutes := parseMinutes(c, 1440)
	limit := 10
	if l, err := strconv.Atoi(c.DefaultQuery("limit", "10")); err == nil && l > 0 && l <= 100 {
		limit = l
	}
	sinceTimestamp := time.Now().Add(-time.Duration(minutes) * time.Minute).Unix()

	data, err := model.GetTopUsers(sinceTimestamp, limit)
	if err != nil {
		common.ApiErrorMsg(c, "获取 Top 用户失败: "+err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    data,
	})
}

// GetMonitorErrors 获取错误分析数据
func GetMonitorErrors(c *gin.Context) {
	minutes := parseMinutes(c, 1440)
	limit := 20
	if l, err := strconv.Atoi(c.DefaultQuery("limit", "20")); err == nil && l > 0 && l <= 100 {
		limit = l
	}
	sinceTimestamp := time.Now().Add(-time.Duration(minutes) * time.Minute).Unix()

	data, err := model.GetMonitorErrors(sinceTimestamp, limit)
	if err != nil {
		common.ApiErrorMsg(c, "获取错误分析数据失败: "+err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    data,
	})
}

// parseMinutes 从查询参数解析分钟数
func parseMinutes(c *gin.Context, defaultMinutes int) int {
	minutes := defaultMinutes
	if m, err := strconv.Atoi(c.DefaultQuery("minutes", strconv.Itoa(defaultMinutes))); err == nil && m > 0 && m <= 14400 {
		minutes = m
	}
	return minutes
}
