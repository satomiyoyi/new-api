package model

import (
	"math"
	"sort"
	"strconv"
	"time"

	"github.com/QuantumNous/new-api/common"
)

// MonitorOverview 实时概览数据
type MonitorOverview struct {
	RPM               int     `json:"rpm"`
	TPM               int     `json:"tpm"`
	ActiveConnections int64   `json:"active_connections"`
	ErrorRate         float64 `json:"error_rate"`
	AvgFirstRespTime  float64 `json:"avg_first_resp_time"`
	TodayQuota        int64   `json:"today_quota"`
	TodayRequests     int64   `json:"today_requests"`
}

// ChannelHealthItem 渠道健康数据
type ChannelHealthItem struct {
	ChannelId    int     `json:"channel_id"`
	ChannelName  string  `json:"channel_name"`
	Requests     int64   `json:"requests"`
	Errors       int64   `json:"errors"`
	SuccessRate  float64 `json:"success_rate"`
	AvgUseTime   float64 `json:"avg_use_time"`
	AvgFRT       float64 `json:"avg_frt"`
	TotalTokens  int64   `json:"total_tokens"`
	TotalQuota   int64   `json:"total_quota"`
}

// ModelPerfItem 模型性能数据
type ModelPerfItem struct {
	ModelName   string  `json:"model_name"`
	Requests    int64   `json:"requests"`
	Errors      int64   `json:"errors"`
	ErrorRate   float64 `json:"error_rate"`
	AvgUseTime  float64 `json:"avg_use_time"`
	AvgFRT      float64 `json:"avg_frt"`
	TotalTokens int64   `json:"total_tokens"`
	TotalQuota  int64   `json:"total_quota"`
}

// TrendPoint 趋势数据点
type TrendPoint struct {
	Timestamp int64 `json:"timestamp"`
	Value     int64 `json:"value"`
}

// FloatTrendPoint 浮点趋势数据点（用于延迟等指标）
type FloatTrendPoint struct {
	Timestamp int64   `json:"timestamp"`
	Value     float64 `json:"value"`
}

// TrendsData 趋势数据
type TrendsData struct {
	Requests []TrendPoint      `json:"requests"`
	Tokens   []TrendPoint      `json:"tokens"`
	Errors   []TrendPoint      `json:"errors"`
	Quota    []TrendPoint      `json:"quota"`
	FRT      []FloatTrendPoint `json:"frt"`
}

// LatencyStats 延迟分布
type LatencyStats struct {
	P50 float64 `json:"p50"`
	P90 float64 `json:"p90"`
	P95 float64 `json:"p95"`
	P99 float64 `json:"p99"`
}

// LatencyData 延迟数据（包含总延迟和首字延迟）
type LatencyData struct {
	UseTime          LatencyStats `json:"use_time"`
	FirstRespTime    LatencyStats `json:"first_resp_time"`
	StreamPercentage float64      `json:"stream_percentage"`
}

// TopUserItem Top 用户
type TopUserItem struct {
	UserId   int    `json:"user_id"`
	Username string `json:"username"`
	Requests int64  `json:"requests"`
	Quota    int64  `json:"quota"`
	Tokens   int64  `json:"tokens"`
}

// ErrorTypeItem 错误类型分布
type ErrorTypeItem struct {
	Content string `json:"content"`
	Count   int64  `json:"count"`
}

// ErrorLogItem 最近错误
type ErrorLogItem struct {
	Id          int    `json:"id"`
	CreatedAt   int64  `json:"created_at"`
	ChannelId   int    `json:"channel_id"`
	ChannelName string `json:"channel_name"`
	ModelName   string `json:"model_name"`
	Content     string `json:"content"`
	Username    string `json:"username"`
}

// ErrorsData 错误分析数据
type ErrorsData struct {
	RecentErrors []ErrorLogItem `json:"recent_errors"`
	TotalErrors  int64          `json:"total_errors"`
}

// GetMonitorOverview 获取监控概览数据
func GetMonitorOverview() (*MonitorOverview, error) {
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).Unix()
	fiveMinAgo := now.Add(-5 * time.Minute).Unix()
	oneMinAgo := now.Add(-1 * time.Minute).Unix()

	overview := &MonitorOverview{}

	// RPM: 最近60秒请求数（consume + error）
	var rpm int64
	LOG_DB.Table("logs").
		Where("created_at >= ? AND (type = ? OR type = ?)", oneMinAgo, LogTypeConsume, LogTypeError).
		Count(&rpm)
	overview.RPM = int(rpm)

	// TPM: 最近60秒 token 总量
	var tpm struct {
		Total *int64
	}
	LOG_DB.Table("logs").
		Select("COALESCE(SUM(prompt_tokens + completion_tokens), 0) as total").
		Where("created_at >= ? AND type = ?", oneMinAgo, LogTypeConsume).
		Scan(&tpm)
	if tpm.Total != nil {
		overview.TPM = int(*tpm.Total)
	}

	// 错误率：最近5分钟
	var totalReqs int64
	var errorReqs int64
	LOG_DB.Table("logs").
		Where("created_at >= ? AND (type = ? OR type = ?)", fiveMinAgo, LogTypeConsume, LogTypeError).
		Count(&totalReqs)
	LOG_DB.Table("logs").
		Where("created_at >= ? AND type = ?", fiveMinAgo, LogTypeError).
		Count(&errorReqs)
	if totalReqs > 0 {
		overview.ErrorRate = math.Round(float64(errorReqs)/float64(totalReqs)*10000) / 100
	}

	// 今日总消耗
	var todayQuota struct {
		Total *int64
	}
	LOG_DB.Table("logs").
		Select("COALESCE(SUM(quota), 0) as total").
		Where("created_at >= ? AND type = ?", todayStart, LogTypeConsume).
		Scan(&todayQuota)
	if todayQuota.Total != nil {
		overview.TodayQuota = *todayQuota.Total
	}

	// 今日总请求数
	var todayReqs int64
	LOG_DB.Table("logs").
		Where("created_at >= ? AND (type = ? OR type = ?)", todayStart, LogTypeConsume, LogTypeError).
		Count(&todayReqs)
	overview.TodayRequests = todayReqs

	// 平均首字延迟（最近5分钟，应用层解析 frt）
	overview.AvgFirstRespTime = getAvgFRT(fiveMinAgo, 0, 0, "")

	return overview, nil
}

// getAvgFRT 从 other JSON 字段提取 frt 平均值（应用层处理，跨数据库兼容）
func getAvgFRT(sinceTimestamp int64, channelId int, userId int, modelName string) float64 {
	type logOther struct {
		Other string `gorm:"column:other"`
	}
	var logs []logOther

	tx := LOG_DB.Table("logs").Select("other").
		Where("created_at >= ? AND type = ?", sinceTimestamp, LogTypeConsume)
	if channelId > 0 {
		tx = tx.Where("channel_id = ?", channelId)
	}
	if userId > 0 {
		tx = tx.Where("user_id = ?", userId)
	}
	if modelName != "" {
		tx = tx.Where("model_name = ?", modelName)
	}
	tx = tx.Where("other != '' AND other != '{}'").
		Limit(5000).
		Find(&logs)

	var sum float64
	var count int
	for _, l := range logs {
		var m map[string]interface{}
		if err := common.UnmarshalJsonStr(l.Other, &m); err != nil {
			continue
		}
		if frt, ok := m["frt"]; ok {
			if frtVal, ok := frt.(float64); ok && frtVal > 0 {
				sum += frtVal
				count++
			}
		}
	}
	if count > 0 {
		return math.Round(sum/float64(count)*100) / 100
	}
	return 0
}

// getFRTValues 获取 frt 值列表
func getFRTValues(sinceTimestamp int64) []float64 {
	type logOther struct {
		Other string `gorm:"column:other"`
	}
	var logs []logOther

	LOG_DB.Table("logs").Select("other").
		Where("created_at >= ? AND type = ?", sinceTimestamp, LogTypeConsume).
		Where("other != '' AND other != '{}'").
		Limit(10000).
		Find(&logs)

	var vals []float64
	for _, l := range logs {
		var m map[string]interface{}
		if err := common.UnmarshalJsonStr(l.Other, &m); err != nil {
			continue
		}
		if frt, ok := m["frt"]; ok {
			if frtVal, ok := frt.(float64); ok && frtVal > 0 {
				vals = append(vals, frtVal)
			}
		}
	}
	return vals
}

// getFRTByBuckets 按时间桶聚合 FRT 平均值（应用层处理，跨数据库兼容）
func getFRTByBuckets(sinceTimestamp int64, granularity int64, channelId int, modelName string) map[int64]float64 {
	type logRow struct {
		CreatedAt int64  `gorm:"column:created_at"`
		Other     string `gorm:"column:other"`
	}
	var logs []logRow

	tx := LOG_DB.Table("logs").Select("created_at, other").
		Where("created_at >= ? AND type = ?", sinceTimestamp, LogTypeConsume).
		Where("other != '' AND other != '{}'")
	if channelId > 0 {
		tx = tx.Where("channel_id = ?", channelId)
	}
	if modelName != "" {
		tx = tx.Where("model_name = ?", modelName)
	}
	tx.Limit(20000).Find(&logs)

	// bucket -> (sum, count)
	type acc struct {
		sum   float64
		count int
	}
	bucketAcc := make(map[int64]*acc)

	for _, l := range logs {
		var m map[string]interface{}
		if err := common.UnmarshalJsonStr(l.Other, &m); err != nil {
			continue
		}
		frt, ok := m["frt"]
		if !ok {
			continue
		}
		frtVal, ok := frt.(float64)
		if !ok || frtVal <= 0 {
			continue
		}
		bucket := (l.CreatedAt / granularity) * granularity
		a, exists := bucketAcc[bucket]
		if !exists {
			a = &acc{}
			bucketAcc[bucket] = a
		}
		a.sum += frtVal
		a.count++
	}

	result := make(map[int64]float64, len(bucketAcc))
	for bucket, a := range bucketAcc {
		if a.count > 0 {
			result[bucket] = math.Round(a.sum/float64(a.count)*100) / 100
		}
	}
	return result
}

// GetChannelHealth 获取渠道健康数据
func GetChannelHealth(sinceTimestamp int64) ([]ChannelHealthItem, error) {
	type channelAgg struct {
		ChannelId int   `gorm:"column:channel_id"`
		Requests  int64 `gorm:"column:requests"`
		AvgTime   float64 `gorm:"column:avg_time"`
		Tokens    int64 `gorm:"column:tokens"`
		Quota     int64 `gorm:"column:quota"`
	}

	var aggs []channelAgg
	err := LOG_DB.Table("logs").
		Select("channel_id, COUNT(*) as requests, AVG(use_time) as avg_time, COALESCE(SUM(prompt_tokens + completion_tokens), 0) as tokens, COALESCE(SUM(quota), 0) as quota").
		Where("created_at >= ? AND (type = ? OR type = ?)", sinceTimestamp, LogTypeConsume, LogTypeError).
		Where("channel_id > 0").
		Group("channel_id").
		Order("requests DESC").
		Limit(100).
		Find(&aggs).Error
	if err != nil {
		return nil, err
	}

	// 查询各渠道错误数
	type channelError struct {
		ChannelId int   `gorm:"column:channel_id"`
		Errors    int64 `gorm:"column:errors"`
	}
	var errs []channelError
	LOG_DB.Table("logs").
		Select("channel_id, COUNT(*) as errors").
		Where("created_at >= ? AND type = ?", sinceTimestamp, LogTypeError).
		Where("channel_id > 0").
		Group("channel_id").
		Find(&errs)

	errMap := make(map[int]int64)
	for _, e := range errs {
		errMap[e.ChannelId] = e.Errors
	}

	// 获取渠道名称
	channelIds := make([]int, 0, len(aggs))
	for _, a := range aggs {
		channelIds = append(channelIds, a.ChannelId)
	}
	nameMap := getChannelNames(channelIds)

	// 组装结果
	result := make([]ChannelHealthItem, 0, len(aggs))
	for _, a := range aggs {
		errors := errMap[a.ChannelId]
		successRate := float64(100)
		if a.Requests > 0 {
			successRate = math.Round(float64(a.Requests-errors)/float64(a.Requests)*10000) / 100
		}
		item := ChannelHealthItem{
			ChannelId:   a.ChannelId,
			ChannelName: nameMap[a.ChannelId],
			Requests:    a.Requests,
			Errors:      errors,
			SuccessRate: successRate,
			AvgUseTime:  math.Round(a.AvgTime*100) / 100,
			TotalTokens: a.Tokens,
			TotalQuota:  a.Quota,
		}
		item.AvgFRT = getAvgFRT(sinceTimestamp, a.ChannelId, 0, "")
		result = append(result, item)
	}

	return result, nil
}

// getChannelNames 批量获取渠道名称
func getChannelNames(ids []int) map[int]string {
	if len(ids) == 0 {
		return nil
	}
	nameMap := make(map[int]string)

	if common.MemoryCacheEnabled {
		for _, id := range ids {
			if ch, err := CacheGetChannel(id); err == nil {
				nameMap[id] = ch.Name
			}
		}
	} else {
		var channels []struct {
			Id   int    `gorm:"column:id"`
			Name string `gorm:"column:name"`
		}
		DB.Table("channels").Select("id, name").Where("id IN ?", ids).Find(&channels)
		for _, ch := range channels {
			nameMap[ch.Id] = ch.Name
		}
	}
	return nameMap
}

// GetModelPerformance 获取模型性能数据
func GetModelPerformance(sinceTimestamp int64) ([]ModelPerfItem, error) {
	type modelAgg struct {
		ModelName string  `gorm:"column:model_name"`
		Requests  int64   `gorm:"column:requests"`
		AvgTime   float64 `gorm:"column:avg_time"`
		Tokens    int64   `gorm:"column:tokens"`
		Quota     int64   `gorm:"column:quota"`
	}

	var aggs []modelAgg
	err := LOG_DB.Table("logs").
		Select("model_name, COUNT(*) as requests, AVG(use_time) as avg_time, COALESCE(SUM(prompt_tokens + completion_tokens), 0) as tokens, COALESCE(SUM(quota), 0) as quota").
		Where("created_at >= ? AND (type = ? OR type = ?)", sinceTimestamp, LogTypeConsume, LogTypeError).
		Where("model_name != ''").
		Group("model_name").
		Order("requests DESC").
		Limit(100).
		Find(&aggs).Error
	if err != nil {
		return nil, err
	}

	// 查询各模型错误数
	type modelError struct {
		ModelName string `gorm:"column:model_name"`
		Errors    int64  `gorm:"column:errors"`
	}
	var errs []modelError
	LOG_DB.Table("logs").
		Select("model_name, COUNT(*) as errors").
		Where("created_at >= ? AND type = ?", sinceTimestamp, LogTypeError).
		Where("model_name != ''").
		Group("model_name").
		Find(&errs)

	errMap := make(map[string]int64)
	for _, e := range errs {
		errMap[e.ModelName] = e.Errors
	}

	result := make([]ModelPerfItem, 0, len(aggs))
	for _, a := range aggs {
		errors := errMap[a.ModelName]
		errRate := float64(0)
		if a.Requests > 0 {
			errRate = math.Round(float64(errors)/float64(a.Requests)*10000) / 100
		}
		item := ModelPerfItem{
			ModelName:   a.ModelName,
			Requests:    a.Requests,
			Errors:      errors,
			ErrorRate:   errRate,
			AvgUseTime:  math.Round(a.AvgTime*100) / 100,
			TotalTokens: a.Tokens,
			TotalQuota:  a.Quota,
		}
		item.AvgFRT = getAvgFRT(sinceTimestamp, 0, 0, a.ModelName)
		result = append(result, item)
	}

	return result, nil
}

// TrendFilter 趋势筛选条件
type TrendFilter struct {
	SinceTimestamp int64
	Granularity    int64
	ChannelId      int
	ModelName      string
}

// GetMonitorTrends 获取趋势数据
func GetMonitorTrends(filter TrendFilter) (*TrendsData, error) {
	granularity := filter.Granularity
	if granularity <= 0 {
		granularity = 60
	}
	sinceTimestamp := filter.SinceTimestamp

	type timeAgg struct {
		Bucket   int64 `gorm:"column:bucket"`
		Requests int64 `gorm:"column:requests"`
		Tokens   int64 `gorm:"column:tokens"`
		Quota    int64 `gorm:"column:quota"`
	}

	// 请求量、Token、Quota 趋势
	var aggs []timeAgg
	bucketExpr := ""
	granStr := strconv.FormatInt(granularity, 10)
	if common.UsingMySQL {
		bucketExpr = "(created_at DIV " + granStr + ") * " + granStr
	} else {
		// PostgreSQL and SQLite both use / for integer division
		bucketExpr = "(created_at / " + granStr + ") * " + granStr
	}

	tx := LOG_DB.Table("logs").
		Select(bucketExpr + " as bucket, COUNT(*) as requests, COALESCE(SUM(prompt_tokens + completion_tokens), 0) as tokens, COALESCE(SUM(quota), 0) as quota").
		Where("created_at >= ? AND (type = ? OR type = ?)", sinceTimestamp, LogTypeConsume, LogTypeError)
	if filter.ChannelId > 0 {
		tx = tx.Where("channel_id = ?", filter.ChannelId)
	}
	if filter.ModelName != "" {
		tx = tx.Where("model_name = ?", filter.ModelName)
	}
	err := tx.Group("bucket").Order("bucket ASC").Find(&aggs).Error
	if err != nil {
		return nil, err
	}

	// 错误趋势
	type errorAgg struct {
		Bucket int64 `gorm:"column:bucket"`
		Errors int64 `gorm:"column:errors"`
	}
	var errAggs []errorAgg
	errTx := LOG_DB.Table("logs").
		Select(bucketExpr + " as bucket, COUNT(*) as errors").
		Where("created_at >= ? AND type = ?", sinceTimestamp, LogTypeError)
	if filter.ChannelId > 0 {
		errTx = errTx.Where("channel_id = ?", filter.ChannelId)
	}
	if filter.ModelName != "" {
		errTx = errTx.Where("model_name = ?", filter.ModelName)
	}
	errTx.Group("bucket").Order("bucket ASC").Find(&errAggs)

	errBucketMap := make(map[int64]int64)
	for _, e := range errAggs {
		errBucketMap[e.Bucket] = e.Errors
	}

	trends := &TrendsData{
		Requests: make([]TrendPoint, 0, len(aggs)),
		Tokens:   make([]TrendPoint, 0, len(aggs)),
		Errors:   make([]TrendPoint, 0, len(aggs)),
		Quota:    make([]TrendPoint, 0, len(aggs)),
		FRT:      nil,
	}

	// 收集所有 bucket 用于 FRT
	buckets := make([]int64, 0, len(aggs))
	for _, a := range aggs {
		trends.Requests = append(trends.Requests, TrendPoint{Timestamp: a.Bucket, Value: a.Requests})
		trends.Tokens = append(trends.Tokens, TrendPoint{Timestamp: a.Bucket, Value: a.Tokens})
		trends.Quota = append(trends.Quota, TrendPoint{Timestamp: a.Bucket, Value: a.Quota})
		trends.Errors = append(trends.Errors, TrendPoint{Timestamp: a.Bucket, Value: errBucketMap[a.Bucket]})
		buckets = append(buckets, a.Bucket)
	}

	// FRT 趋势：从 other JSON 字段提取，应用层按时间桶聚合
	frtBucketMap := getFRTByBuckets(sinceTimestamp, granularity, filter.ChannelId, filter.ModelName)
	frtPoints := make([]FloatTrendPoint, 0, len(buckets))
	for _, b := range buckets {
		frtPoints = append(frtPoints, FloatTrendPoint{Timestamp: b, Value: frtBucketMap[b]})
	}
	trends.FRT = frtPoints

	return trends, nil
}

// GetMonitorLatency 获取延迟分布
func GetMonitorLatency(sinceTimestamp int64) (*LatencyData, error) {
	// 获取 use_time 值
	var useTimes []float64
	LOG_DB.Table("logs").
		Select("use_time").
		Where("created_at >= ? AND type = ? AND use_time > 0", sinceTimestamp, LogTypeConsume).
		Order("use_time ASC").
		Limit(10000).
		Pluck("use_time", &useTimes)

	// 获取 FRT 值
	frtValues := getFRTValues(sinceTimestamp)

	// 获取流式百分比
	var totalCount int64
	var streamCount int64
	LOG_DB.Table("logs").
		Where("created_at >= ? AND type = ?", sinceTimestamp, LogTypeConsume).
		Count(&totalCount)
	LOG_DB.Table("logs").
		Where("created_at >= ? AND type = ? AND is_stream = ?", sinceTimestamp, LogTypeConsume, true).
		Count(&streamCount)

	streamPct := float64(0)
	if totalCount > 0 {
		streamPct = math.Round(float64(streamCount)/float64(totalCount)*10000) / 100
	}

	return &LatencyData{
		UseTime:          calcPercentiles(useTimes),
		FirstRespTime:    calcPercentiles(frtValues),
		StreamPercentage: streamPct,
	}, nil
}

// calcPercentiles 计算百分位数
func calcPercentiles(vals []float64) LatencyStats {
	if len(vals) == 0 {
		return LatencyStats{}
	}
	sort.Float64s(vals)
	n := len(vals)

	pIdx := func(p float64) int {
		idx := int(math.Ceil(p*float64(n))) - 1
		if idx < 0 {
			return 0
		}
		if idx >= n {
			return n - 1
		}
		return idx
	}

	return LatencyStats{
		P50: math.Round(vals[pIdx(0.50)]*100) / 100,
		P90: math.Round(vals[pIdx(0.90)]*100) / 100,
		P95: math.Round(vals[pIdx(0.95)]*100) / 100,
		P99: math.Round(vals[pIdx(0.99)]*100) / 100,
	}
}

// GetTopUsers 获取 Top N 用户
func GetTopUsers(sinceTimestamp int64, limit int) ([]TopUserItem, error) {
	if limit <= 0 {
		limit = 10
	}

	type userAgg struct {
		UserId   int    `gorm:"column:user_id"`
		Username string `gorm:"column:username"`
		Requests int64  `gorm:"column:requests"`
		Quota    int64  `gorm:"column:quota"`
		Tokens   int64  `gorm:"column:tokens"`
	}

	var aggs []userAgg
	err := LOG_DB.Table("logs").
		Select("user_id, username, COUNT(*) as requests, COALESCE(SUM(quota), 0) as quota, COALESCE(SUM(prompt_tokens + completion_tokens), 0) as tokens").
		Where("created_at >= ? AND (type = ? OR type = ?)", sinceTimestamp, LogTypeConsume, LogTypeError).
		Group("user_id, username").
		Order("quota DESC").
		Limit(limit).
		Find(&aggs).Error
	if err != nil {
		return nil, err
	}

	result := make([]TopUserItem, 0, len(aggs))
	for _, a := range aggs {
		result = append(result, TopUserItem{
			UserId:   a.UserId,
			Username: a.Username,
			Requests: a.Requests,
			Quota:    a.Quota,
			Tokens:   a.Tokens,
		})
	}
	return result, nil
}

// GetMonitorErrors 获取错误分析数据
func GetMonitorErrors(sinceTimestamp int64, limit int) (*ErrorsData, error) {
	if limit <= 0 {
		limit = 20
	}

	// 最近错误日志
	var logs []Log
	err := LOG_DB.Table("logs").
		Where("type = ? AND created_at >= ?", LogTypeError, sinceTimestamp).
		Order("created_at DESC").
		Limit(limit).
		Find(&logs).Error
	if err != nil {
		return nil, err
	}

	// 获取渠道名称
	channelIds := make([]int, 0)
	for _, l := range logs {
		if l.ChannelId > 0 {
			channelIds = append(channelIds, l.ChannelId)
		}
	}
	nameMap := getChannelNames(channelIds)

	recentErrors := make([]ErrorLogItem, 0, len(logs))
	for _, l := range logs {
		recentErrors = append(recentErrors, ErrorLogItem{
			Id:          l.Id,
			CreatedAt:   l.CreatedAt,
			ChannelId:   l.ChannelId,
			ChannelName: nameMap[l.ChannelId],
			ModelName:   l.ModelName,
			Content:     l.Content,
			Username:    l.Username,
		})
	}

	// 总错误数
	var totalErrors int64
	LOG_DB.Table("logs").
		Where("type = ? AND created_at >= ?", LogTypeError, sinceTimestamp).
		Count(&totalErrors)

	return &ErrorsData{
		RecentErrors: recentErrors,
		TotalErrors:  totalErrors,
	}, nil
}
