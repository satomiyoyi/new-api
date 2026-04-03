import React from 'react';
import { Card, Spin, Table, Tag } from '@douyinfe/semi-ui';
import { Package } from 'lucide-react';
import { renderQuota } from '../../helpers/render';

const ModelPerformance = ({ models, loading, t }) => {
  const columns = [
    {
      title: t('模型'),
      dataIndex: 'model_name',
      key: 'model_name',
      width: 150,
      ellipsis: true,
    },
    {
      title: t('请求量'),
      dataIndex: 'requests',
      key: 'requests',
      sorter: (a, b) => a.requests - b.requests,
      width: 80,
    },
    {
      title: t('错误率'),
      dataIndex: 'error_rate',
      key: 'error_rate',
      sorter: (a, b) => a.error_rate - b.error_rate,
      render: (v) => {
        const color = v <= 1 ? 'green' : v <= 5 ? 'yellow' : 'red';
        return <Tag color={color} size='small'>{v}%</Tag>;
      },
      width: 80,
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
      <div className='flex items-center gap-2'><Package size={16} />{t('模型性能')}</div>
    } className='!rounded-xl'>
      {loading && models.length === 0 ? (
        <div className='flex justify-center py-8'><Spin /></div>
      ) : (
        <Table
          columns={columns}
          dataSource={models}
          rowKey='model_name'
          pagination={false}
          size='small'
          scroll={{ y: 320 }}
          empty={<div className='text-center text-gray-400 py-4'>{t('暂无数据')}</div>}
        />
      )}
    </Card>
  );
};

export default ModelPerformance;
