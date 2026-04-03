import React from 'react';
import { Card, Spin } from '@douyinfe/semi-ui';
import {
  Zap, Clock, Wifi, AlertTriangle, Timer, DollarSign, Hash,
} from 'lucide-react';
import { renderQuota } from '../../helpers/render';

const cardStyle = {
  display: 'flex',
  flexDirection: 'column',
  alignItems: 'center',
  justifyContent: 'center',
  padding: '16px 12px',
  minHeight: 100,
};

const OverviewCards = ({ overview, loading, t }) => {
  if (loading && !overview) {
    return (
      <div className='grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-7 gap-3 mb-4'>
        {Array.from({ length: 7 }).map((_, i) => (
          <Card key={i} shadows='' bordered className='!rounded-xl'>
            <div style={cardStyle}>
              <Spin />
            </div>
          </Card>
        ))}
      </div>
    );
  }

  if (!overview) return null;

  const items = [
    {
      icon: <Zap size={18} className='text-blue-500' />,
      label: t('RPM'),
      value: overview.rpm,
    },
    {
      icon: <Hash size={18} className='text-green-500' />,
      label: t('TPM'),
      value: formatNumber(overview.tpm),
    },
    {
      icon: <Wifi size={18} className='text-cyan-500' />,
      label: t('活跃连接'),
      value: overview.active_connections,
    },
    {
      icon: <AlertTriangle size={18} className='text-red-500' />,
      label: t('错误率'),
      value: overview.error_rate + '%',
    },
    {
      icon: <Timer size={18} className='text-orange-500' />,
      label: t('首字延迟'),
      value: overview.avg_first_resp_time > 0 ? overview.avg_first_resp_time + 'ms' : '-',
    },
    {
      icon: <DollarSign size={18} className='text-purple-500' />,
      label: t('今日消耗'),
      value: renderQuota(overview.today_quota, 2),
    },
    {
      icon: <Clock size={18} className='text-indigo-500' />,
      label: t('今日请求'),
      value: formatNumber(overview.today_requests),
    },
  ];

  return (
    <div className='grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-7 gap-3 mb-4'>
      {items.map((item, idx) => (
        <Card key={idx} shadows='' bordered className='!rounded-xl'>
          <div style={cardStyle}>
            <div className='mb-1'>{item.icon}</div>
            <div className='text-xs text-gray-500 mb-1'>{item.label}</div>
            <div className='text-lg font-bold'>{item.value}</div>
          </div>
        </Card>
      ))}
    </div>
  );
};

function formatNumber(n) {
  if (n === null || n === undefined) return '-';
  if (n >= 1000000) return (n / 1000000).toFixed(1) + 'M';
  if (n >= 1000) return (n / 1000).toFixed(1) + 'K';
  return String(n);
}

export default OverviewCards;
