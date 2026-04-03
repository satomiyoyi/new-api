import React, { useMemo } from 'react';
import { Card, Spin, Tag } from '@douyinfe/semi-ui';
import { BarChart3 } from 'lucide-react';
import { VChart } from '@visactor/react-vchart';

const CHART_CONFIG = { mode: 'desktop-browser' };

const LatencyPanel = ({ latency, loading, t }) => {
  const spec = useMemo(() => {
    if (!latency) return null;

    const data = [
      { percentile: 'P50', type: t('总延迟') + '(s)', value: latency.use_time.p50 },
      { percentile: 'P90', type: t('总延迟') + '(s)', value: latency.use_time.p90 },
      { percentile: 'P95', type: t('总延迟') + '(s)', value: latency.use_time.p95 },
      { percentile: 'P99', type: t('总延迟') + '(s)', value: latency.use_time.p99 },
      { percentile: 'P50', type: t('首字延迟') + '(ms)', value: latency.first_resp_time.p50 },
      { percentile: 'P90', type: t('首字延迟') + '(ms)', value: latency.first_resp_time.p90 },
      { percentile: 'P95', type: t('首字延迟') + '(ms)', value: latency.first_resp_time.p95 },
      { percentile: 'P99', type: t('首字延迟') + '(ms)', value: latency.first_resp_time.p99 },
    ];

    return {
      type: 'bar',
      data: [{ id: 'data', values: data }],
      xField: 'percentile',
      yField: 'value',
      seriesField: 'type',
      legends: { visible: true, orient: 'top' },
      axes: [
        { orient: 'bottom' },
        { orient: 'left' },
      ],
      tooltip: { mark: {} },
      animationAppear: false,
    };
  }, [latency, t]);

  return (
    <Card shadows='' bordered headerLine title={
      <div className='flex items-center justify-between w-full'>
        <div className='flex items-center gap-2'>
          <BarChart3 size={16} />
          {t('延迟分布')}
        </div>
        {latency && (
          <Tag size='small' color='blue'>
            {t('流式')} {latency.stream_percentage}%
          </Tag>
        )}
      </div>
    } className='!rounded-xl'>
      <div className='h-64'>
        {loading && !latency ? (
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

export default LatencyPanel;
