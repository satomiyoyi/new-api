import React, { useState, useMemo } from 'react';
import { useTranslation } from 'react-i18next';
import { Select, Button } from '@douyinfe/semi-ui';
import { RefreshCw, Activity } from 'lucide-react';
import { useMonitorData } from '../../hooks/monitor/useMonitorData';
import OverviewCards from './OverviewCards';
import SystemPanel from './SystemPanel';
import ChannelHealth from './ChannelHealth';
import ModelPerformance from './ModelPerformance';
import TrendsPanel from './TrendsPanel';
import LatencyPanel from './LatencyPanel';
import ErrorsPanel from './ErrorsPanel';

const TIME_RANGE_OPTIONS = [
  { value: 5, label: '5 min' },
  { value: 15, label: '15 min' },
  { value: 60, label: '1 hour' },
  { value: 360, label: '6 hours' },
  { value: 1440, label: '24 hours' },
];

const REFRESH_OPTIONS = [
  { value: 0, label: 'Off' },
  { value: 15, label: '15s' },
  { value: 30, label: '30s' },
  { value: 60, label: '60s' },
];

const MonitorPanel = () => {
  const { t } = useTranslation();
  const [minutes, setMinutes] = useState(60);
  const { data, loading, refresh, refreshInterval, setRefreshInterval } = useMonitorData(minutes);

  const timeRangeOpts = useMemo(() =>
    TIME_RANGE_OPTIONS.map(o => ({ ...o, label: t(o.label) || o.label })),
  [t]);

  const refreshOpts = useMemo(() =>
    REFRESH_OPTIONS.map(o => ({ ...o, label: o.value === 0 ? t('关闭') : o.label })),
  [t]);

  return (
    <div className='h-full'>
      {/* Header */}
      <div className='flex flex-col sm:flex-row items-start sm:items-center justify-between mb-4 gap-2'>
        <div className='flex items-center gap-2'>
          <Activity size={20} />
          <span className='text-lg font-semibold'>{t('监控面板')}</span>
        </div>
        <div className='flex items-center gap-2 flex-wrap'>
          <Select
            size='small'
            value={minutes}
            onChange={setMinutes}
            optionList={timeRangeOpts}
            style={{ width: 120 }}
            prefix={<span className='text-xs text-gray-500'>{t('时间范围')}</span>}
          />
          <Select
            size='small'
            value={refreshInterval}
            onChange={setRefreshInterval}
            optionList={refreshOpts}
            style={{ width: 110 }}
            prefix={<span className='text-xs text-gray-500'>{t('刷新')}</span>}
          />
          <Button
            size='small'
            icon={<RefreshCw size={14} className={loading ? 'animate-spin' : ''} />}
            onClick={refresh}
            loading={loading}
          >
            {t('刷新')}
          </Button>
        </div>
      </div>

      {/* Overview Cards */}
      <OverviewCards overview={data.overview} loading={loading} t={t} />

      {/* Trends - self-managed data fetching with own filters */}
      <TrendsPanel channels={data.channels} models={data.models} loading={loading} t={t} />

      {/* Channel Health + Model Performance */}
      <div className='grid grid-cols-1 lg:grid-cols-2 gap-4 mb-4'>
        <ChannelHealth channels={data.channels} loading={loading} t={t} />
        <ModelPerformance models={data.models} loading={loading} t={t} />
      </div>

      {/* System + Latency */}
      <div className='grid grid-cols-1 lg:grid-cols-2 gap-4 mb-4'>
        <SystemPanel system={data.system} loading={loading} t={t} />
        <LatencyPanel latency={data.latency} loading={loading} t={t} />
      </div>

      {/* Errors */}
      <ErrorsPanel errors={data.errors} loading={loading} t={t} />
    </div>
  );
};

export default MonitorPanel;
