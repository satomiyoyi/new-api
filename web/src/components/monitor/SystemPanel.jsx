import React from 'react';
import { Card, Spin, Progress } from '@douyinfe/semi-ui';
import { Cpu, HardDrive, MemoryStick, Server } from 'lucide-react';

const SystemPanel = ({ system, loading, t }) => {
  if (loading && !system) {
    return (
      <Card shadows='' bordered headerLine title={t('系统资源')} className='!rounded-xl'>
        <div className='flex justify-center py-8'><Spin /></div>
      </Card>
    );
  }

  if (!system) return null;

  const resources = [
    { icon: <Cpu size={16} />, label: 'CPU', value: system.cpu_usage },
    { icon: <MemoryStick size={16} />, label: t('内存'), value: system.memory_usage },
    { icon: <HardDrive size={16} />, label: t('磁盘'), value: system.disk_usage },
  ];

  const getColor = (v) => {
    if (v >= 90) return 'var(--semi-color-danger)';
    if (v >= 70) return 'var(--semi-color-warning)';
    return 'var(--semi-color-success)';
  };

  return (
    <Card shadows='' bordered headerLine title={
      <div className='flex items-center gap-2'><Server size={16} />{t('系统资源')}</div>
    } className='!rounded-xl'>
      <div className='space-y-4'>
        {resources.map((r, idx) => (
          <div key={idx} className='flex items-center gap-3'>
            <div className='flex items-center gap-2 w-16 text-sm'>
              {r.icon}
              <span>{r.label}</span>
            </div>
            <div className='flex-1'>
              <Progress
                percent={Math.round(r.value * 100) / 100}
                stroke={getColor(r.value)}
                showInfo
                size='large'
              />
            </div>
          </div>
        ))}
        <div className='grid grid-cols-2 gap-2 text-xs text-gray-500 mt-2'>
          <div>Goroutines: <span className='font-mono font-medium text-gray-700'>{system.num_goroutine}</span></div>
          <div>Go Alloc: <span className='font-mono font-medium text-gray-700'>{formatBytes(system.go_alloc)}</span></div>
          <div>Go Sys: <span className='font-mono font-medium text-gray-700'>{formatBytes(system.go_sys)}</span></div>
          <div>GC: <span className='font-mono font-medium text-gray-700'>{system.go_num_gc}</span></div>
        </div>
      </div>
    </Card>
  );
};

function formatBytes(bytes) {
  if (!bytes) return '-';
  if (bytes >= 1073741824) return (bytes / 1073741824).toFixed(1) + ' GB';
  if (bytes >= 1048576) return (bytes / 1048576).toFixed(1) + ' MB';
  if (bytes >= 1024) return (bytes / 1024).toFixed(1) + ' KB';
  return bytes + ' B';
}

export default SystemPanel;
