import React, { useState, useMemo, useEffect, useCallback } from 'react';
import { Card, Spin, Tabs, TabPane, Select } from '@douyinfe/semi-ui';
import { TrendingUp } from 'lucide-react';
import { VChart } from '@visactor/react-vchart';
import { API } from '../../helpers/api';

const CHART_CONFIG = { mode: 'desktop-browser' };

const TIME_OPTIONS = [
  { value: 60, label: '1h' },
  { value: 360, label: '6h' },
  { value: 1440, label: '24h' },
  { value: 4320, label: '3d' },
  { value: 10080, label: '7d' },
];

const TrendsPanel = ({ channels, models, loading: parentLoading, t }) => {
  const [tab, setTab] = useState('requests');
  const [minutes, setMinutes] = useState(1440);
  const [channelId, setChannelId] = useState(0);
  const [modelName, setModelName] = useState('');
  const [trends, setTrends] = useState(null);
  const [loading, setLoading] = useState(false);

  const granularity = useMemo(() => {
    if (minutes <= 60) return 60;
    if (minutes <= 360) return 300;
    if (minutes <= 1440) return 600;
    if (minutes <= 4320) return 1800;
    return 3600;
  }, [minutes]);

  const fetchTrends = useCallback(async () => {
    setLoading(true);
    try {
      const params = { minutes, granularity };
      if (channelId > 0) params.channel_id = channelId;
      if (modelName) params.model_name = modelName;
      const res = await API.get('/api/monitor/trends', { params, disableDuplicate: true });
      if (res.data.success) {
        setTrends(res.data.data);
      }
    } catch (e) {
      console.error('Fetch trends error:', e);
    } finally {
      setLoading(false);
    }
  }, [minutes, granularity, channelId, modelName]);

  useEffect(() => {
    fetchTrends();
  }, [fetchTrends]);

  const channelOptions = useMemo(() => {
    const opts = [{ value: 0, label: t('全部渠道') }];
    if (channels && channels.length > 0) {
      channels.forEach(ch => {
        opts.push({
          value: ch.channel_id,
          label: ch.channel_name || `#${ch.channel_id}`,
        });
      });
    }
    return opts;
  }, [channels, t]);

  const modelOptions = useMemo(() => {
    const opts = [{ value: '', label: t('全部模型') }];
    if (models && models.length > 0) {
      models.forEach(m => {
        opts.push({ value: m.model_name, label: m.model_name });
      });
    }
    return opts;
  }, [models, t]);

  const spec = useMemo(() => {
    if (!trends) return null;

    const dataMap = {
      requests: trends.requests,
      tokens: trends.tokens,
      errors: trends.errors,
      quota: trends.quota,
      frt: trends.frt,
    };

    const labelMap = {
      requests: t('请求量'),
      tokens: 'Tokens',
      errors: t('错误数'),
      quota: t('额度'),
      frt: t('首字延迟') + ' (ms)',
    };

    const colorMap = {
      requests: '#3b82f6',
      tokens: '#10b981',
      errors: '#ef4444',
      quota: '#8b5cf6',
      frt: '#f97316',
    };

    const points = (dataMap[tab] || []).map(p => ({
      time: formatTime(p.timestamp, minutes),
      value: p.value,
    }));

    if (points.length === 0) return null;

    return {
      type: 'area',
      data: [{ id: 'data', values: points }],
      xField: 'time',
      yField: 'value',
      point: { visible: points.length <= 30, style: { size: 4 } },
      line: { style: { stroke: colorMap[tab], lineWidth: 2 } },
      area: {
        style: {
          fill: {
            gradient: 'linear',
            x0: 0, y0: 0, x1: 0, y1: 1,
            stops: [
              { offset: 0, color: colorMap[tab] + '40' },
              { offset: 1, color: colorMap[tab] + '05' },
            ],
          },
        },
      },
      axes: [
        { orient: 'bottom', label: { style: { fontSize: 10 }, autoRotate: true } },
        { orient: 'left', label: { style: { fontSize: 10 } } },
      ],
      tooltip: { mark: { title: { value: labelMap[tab] } } },
      crosshair: { xField: { visible: true } },
      animationAppear: false,
    };
  }, [trends, tab, t, minutes]);

  return (
    <Card
      shadows='' bordered headerLine
      className='!rounded-xl mb-4'
      title={
        <div className='flex flex-col gap-2'>
          <div className='flex flex-col sm:flex-row sm:items-center sm:justify-between w-full gap-2'>
            <div className='flex items-center gap-2'>
              <TrendingUp size={16} />
              {t('趋势')}
            </div>
            <Tabs type='slash' activeKey={tab} onChange={setTab} size='small'>
              <TabPane tab={t('请求量')} itemKey='requests' />
              <TabPane tab='Tokens' itemKey='tokens' />
              <TabPane tab={t('错误数')} itemKey='errors' />
              <TabPane tab={t('额度')} itemKey='quota' />
              <TabPane tab={t('首字延迟')} itemKey='frt' />
            </Tabs>
          </div>
          <div className='flex items-center gap-2 flex-wrap'>
            <Select
              size='small'
              value={minutes}
              onChange={setMinutes}
              optionList={TIME_OPTIONS}
              style={{ width: 90 }}
            />
            <Select
              size='small'
              value={channelId}
              onChange={setChannelId}
              optionList={channelOptions}
              style={{ width: 150 }}
              filter
            />
            <Select
              size='small'
              value={modelName}
              onChange={setModelName}
              optionList={modelOptions}
              style={{ width: 200 }}
              filter
            />
          </div>
        </div>
      }
      bodyStyle={{ padding: 0 }}
    >
      <div className='h-72 p-2'>
        {loading ? (
          <div className='flex justify-center items-center h-full'><Spin /></div>
        ) : spec ? (
          <VChart spec={spec} option={CHART_CONFIG} />
        ) : (
          <div className='flex justify-center items-center h-full text-gray-400'>
            {t('暂无数据')}
          </div>
        )}
      </div>
    </Card>
  );
};

function formatTime(ts, minutes) {
  const d = new Date(ts * 1000);
  if (minutes > 1440) {
    return `${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')} ${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`;
  }
  return `${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`;
}

export default TrendsPanel;
