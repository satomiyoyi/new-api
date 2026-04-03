import React from 'react';
import { Card, Spin, Table, Tag } from '@douyinfe/semi-ui';
import { AlertCircle } from 'lucide-react';

const ErrorsPanel = ({ errors, loading, t }) => {
  const columns = [
    {
      title: t('时间'),
      dataIndex: 'created_at',
      key: 'created_at',
      render: (v) => new Date(v * 1000).toLocaleString(),
      width: 170,
    },
    {
      title: t('渠道'),
      dataIndex: 'channel_name',
      key: 'channel_name',
      render: (text, record) => text || (record.channel_id > 0 ? `#${record.channel_id}` : '-'),
      width: 100,
    },
    {
      title: t('模型'),
      dataIndex: 'model_name',
      key: 'model_name',
      width: 150,
      ellipsis: true,
    },
    {
      title: t('用户'),
      dataIndex: 'username',
      key: 'username',
      width: 80,
    },
    {
      title: t('错误内容'),
      dataIndex: 'content',
      key: 'content',
      ellipsis: true,
    },
  ];

  const recentErrors = errors?.recent_errors || [];
  const totalErrors = errors?.total_errors || 0;

  return (
    <Card shadows='' bordered headerLine title={
      <div className='flex items-center justify-between w-full'>
        <div className='flex items-center gap-2'>
          <AlertCircle size={16} />
          {t('最近错误')}
        </div>
        {totalErrors > 0 && (
          <Tag size='small' color='red'>
            {t('共')} {totalErrors} {t('条错误')}
          </Tag>
        )}
      </div>
    } className='!rounded-xl mb-4'>
      {loading && recentErrors.length === 0 ? (
        <div className='flex justify-center py-8'><Spin /></div>
      ) : (
        <Table
          columns={columns}
          dataSource={recentErrors}
          rowKey='id'
          pagination={false}
          size='small'
          scroll={{ y: 280 }}
          empty={<div className='text-center text-gray-400 py-4'>{t('暂无错误')}</div>}
        />
      )}
    </Card>
  );
};

export default ErrorsPanel;
