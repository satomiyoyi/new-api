import React from 'react';
import { Card, Spin, Table, Tag } from '@douyinfe/semi-ui';
import { Layers } from 'lucide-react';
import { renderQuota } from '../../helpers/render';

const ChannelHealth = ({ channels, loading, t }) => {
  const columns = [
    {
      title: t('渠道'),
      dataIndex: 'channel_name',
      key: 'channel_name',
      render: (text, record) => text || `#${record.channel_id}`,
      width: 120,
    },
    {
      title: t('成功率'),
      dataIndex: 'success_rate',
      key: 'success_rate',
      sorter: (a, b) => a.success_rate - b.success_rate,
      render: (v) => {
        const color = v >= 99 ? 'green' : v >= 95 ? 'yellow' : 'red';
        return <Tag color={color} size='small'>{v}%</Tag>;
      },
      width: 90,
    },
    {
      title: t('请求量'),
      dataIndex: 'requests',
      key: 'requests',
      sorter: (a, b) => a.requests - b.requests,
      width: 80,
    },
    {
      title: t('错误数'),
      dataIndex: 'errors',
      key: 'errors',
      sorter: (a, b) => a.errors - b.errors,
      render: (v) => v > 0 ? <span className='text-red-500 font-medium'>{v}</span> : '0',
      width: 70,
    },
    {
      title: t('平均延迟'),
      dataIndex: 'avg_use_time',
      key: 'avg_use_time',
      sorter: (a, b) => a.avg_use_time - b.avg_use_time,
      render: (v) => v > 0 ? v + 's' : '-',
      width: 90,
    },
    {
      title: t('首字延迟'),
      dataIndex: 'avg_frt',
      key: 'avg_frt',
      render: (v) => v > 0 ? v + 'ms' : '-',
      width: 90,
    },
    {
      title: t('额度消耗'),
      dataIndex: 'total_quota',
      key: 'total_quota',
      sorter: (a, b) => a.total_quota - b.total_quota,
      render: (v) => renderQuota(v, 2),
      width: 100,
    },
  ];

  return (
    <Card shadows='' bordered headerLine title={
      <div className='flex items-center gap-2'><Layers size={16} />{t('渠道健康')}</div>
    } className='!rounded-xl'>
      {loading && channels.length === 0 ? (
        <div className='flex justify-center py-8'><Spin /></div>
      ) : (
        <Table
          columns={columns}
          dataSource={channels}
          rowKey='channel_id'
          pagination={false}
          size='small'
          scroll={{ y: 320 }}
          empty={<div className='text-center text-gray-400 py-4'>{t('暂无数据')}</div>}
        />
      )}
    </Card>
  );
};

export default ChannelHealth;
