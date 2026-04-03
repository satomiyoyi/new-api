import { useState, useEffect, useCallback, useRef } from 'react';
import { API } from '../../helpers/api';

const defaultData = {
  overview: null,
  system: null,
  channels: [],
  models: [],
  latency: null,
  topUsers: [],
  errors: null,
};

export function useMonitorData(minutes = 60) {
  const [data, setData] = useState(defaultData);
  const [loading, setLoading] = useState(true);
  const timerRef = useRef(null);
  const [refreshInterval, setRefreshInterval] = useState(30);

  const fetchAll = useCallback(async () => {
    try {
      const params = { minutes };

      const [
        overviewRes,
        systemRes,
        channelsRes,
        modelsRes,
        latencyRes,
        topUsersRes,
        errorsRes,
      ] = await Promise.allSettled([
        API.get('/api/monitor/overview'),
        API.get('/api/monitor/system'),
        API.get('/api/monitor/channels', { params }),
        API.get('/api/monitor/models', { params }),
        API.get('/api/monitor/latency', { params }),
        API.get('/api/monitor/top_users', { params: { minutes: 1440, limit: 10 } }),
        API.get('/api/monitor/errors', { params: { minutes: 1440, limit: 20 } }),
      ]);

      setData({
        overview: overviewRes.status === 'fulfilled' && overviewRes.value.data.success ? overviewRes.value.data.data : null,
        system: systemRes.status === 'fulfilled' && systemRes.value.data.success ? systemRes.value.data.data : null,
        channels: channelsRes.status === 'fulfilled' && channelsRes.value.data.success ? channelsRes.value.data.data : [],
        models: modelsRes.status === 'fulfilled' && modelsRes.value.data.success ? modelsRes.value.data.data : [],
        latency: latencyRes.status === 'fulfilled' && latencyRes.value.data.success ? latencyRes.value.data.data : null,
        topUsers: topUsersRes.status === 'fulfilled' && topUsersRes.value.data.success ? topUsersRes.value.data.data : [],
        errors: errorsRes.status === 'fulfilled' && errorsRes.value.data.success ? errorsRes.value.data.data : null,
      });
    } catch (e) {
      console.error('Monitor fetch error:', e);
    } finally {
      setLoading(false);
    }
  }, [minutes]);

  useEffect(() => {
    setLoading(true);
    fetchAll();
  }, [fetchAll]);

  useEffect(() => {
    if (timerRef.current) clearInterval(timerRef.current);
    if (refreshInterval > 0) {
      timerRef.current = setInterval(fetchAll, refreshInterval * 1000);
    }
    return () => {
      if (timerRef.current) clearInterval(timerRef.current);
    };
  }, [fetchAll, refreshInterval]);

  return { data, loading, refresh: fetchAll, refreshInterval, setRefreshInterval };
}
