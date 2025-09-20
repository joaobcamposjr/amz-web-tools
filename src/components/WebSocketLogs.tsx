'use client';

import React, { useEffect, useState, useRef } from 'react';

interface LogMessage {
  type: string;
  timestamp: string;
  level: 'info' | 'success' | 'warning' | 'error';
  step: string;
  message: string;
  process_id?: string;
}

interface WebSocketLogsProps {
  processId?: string | null;
  onLogReceived?: (log: LogMessage) => void;
  processResult?: any; // Resultado do processamento XML
}

const WebSocketLogs: React.FC<WebSocketLogsProps> = ({ processId, onLogReceived, processResult }) => {
  const [logs, setLogs] = useState<LogMessage[]>([]);
  const [isConnected, setIsConnected] = useState(false);
  const [connectionStatus, setConnectionStatus] = useState('Desconectado');
  const pollingIntervalRef = useRef<NodeJS.Timeout | null>(null);
  const logsContainerRef = useRef<HTMLDivElement>(null);

  // Função para fazer scroll automático para a última mensagem
  const scrollToBottom = () => {
    if (logsContainerRef.current) {
      logsContainerRef.current.scrollTop = logsContainerRef.current.scrollHeight;
    }
  };

  // Scroll automático sempre que os logs mudarem
  useEffect(() => {
    scrollToBottom();
  }, [logs]);

  const fetchLogs = async () => {
    if (!processId) return;
    
    try {
      // Se temos o resultado do processamento, usar os logs reais do backend
      let logsArray = null;
      
      if (processResult && processResult.data && processResult.data.logs) {
        logsArray = processResult.data.logs;
      } else if (processResult && processResult.logs) {
        logsArray = processResult.logs;
      }

      if (logsArray && Array.isArray(logsArray)) {
        const realLogs: LogMessage[] = logsArray.map((log: any) => ({
          type: 'log',
          timestamp: log.timestamp || new Date().toISOString(),
          level: log.level || 'info',
          step: log.step || 'Processamento',
          message: log.message || 'Processando...',
          process_id: processId,
        }));

        setLogs(realLogs);
        setIsConnected(true);
        setConnectionStatus('Conectado');
        
        if (onLogReceived) {
          realLogs.forEach(log => onLogReceived(log));
        }
        return;
      }

      // Debug: verificar estrutura do processResult
      console.log('🔍 Debug processResult:', processResult);
      if (processResult) {
        console.log('🔍 processResult.data:', processResult.data);
        console.log('🔍 processResult.data.logs:', processResult.data?.logs);
        console.log('🔍 processResult.logs:', processResult.logs);
      }

      // Se não temos logs do backend, mostrar mensagem de aguardando
      setLogs([]);
      setIsConnected(false);
      setConnectionStatus('Aguardando logs do backend...');
      
    } catch (error) {
      console.error('❌ Erro ao buscar logs:', error);
      setConnectionStatus('Erro de conexão');
      setIsConnected(false);
    }
  };

  useEffect(() => {
    if (processId) {
      fetchLogs();
      
      // Buscar logs a cada 5 segundos apenas se não temos processResult
      if (!processResult) {
        pollingIntervalRef.current = setInterval(() => {
          fetchLogs();
        }, 5000);
      }
    }

    return () => {
      if (pollingIntervalRef.current) {
        clearInterval(pollingIntervalRef.current);
      }
    };
  }, [processId, processResult]);

  const getLevelIcon = (level: string) => {
    switch (level) {
      case 'info':
        return 'ℹ️';
      case 'success':
        return '✅';
      case 'warning':
        return '⚠️';
      case 'error':
        return '❌';
      default:
        return '📝';
    }
  };

  const getLevelColor = (level: string) => {
    switch (level) {
      case 'info':
        return 'text-blue-600';
      case 'success':
        return 'text-green-600';
      case 'warning':
        return 'text-yellow-600';
      case 'error':
        return 'text-red-600';
      default:
        return 'text-gray-600';
    }
  };

  const formatTimestamp = (timestamp: string) => {
    try {
      return new Date(timestamp).toLocaleTimeString('pt-BR');
    } catch {
      return timestamp;
    }
  };

  return (
    <div className="bg-white rounded-lg shadow-md p-6">
      <div className="flex items-center justify-between mb-4">
        <h3 className="text-lg font-semibold text-gray-800">
          Logs do Processo em Tempo Real
        </h3>
        <div className="flex items-center space-x-2">
          <div className={`w-3 h-3 rounded-full ${isConnected ? 'bg-green-500' : 'bg-red-500'}`}></div>
          <span className="text-sm text-gray-600">{connectionStatus}</span>
        </div>
      </div>

      {logs.length === 0 ? (
        <div className="text-center py-8 text-gray-500">
          <div className="text-4xl mb-2">📡</div>
          <p>Aguardando logs...</p>
          <p className="text-sm">Conecte-se ao processo XML para ver os logs em tempo real</p>
        </div>
      ) : (
        <div ref={logsContainerRef} className="space-y-3 max-h-96 overflow-y-auto">
          {logs.map((log, index) => (
            <div
              key={index}
              className="p-3 border-l-4 border-gray-300"
            >
              <div className="flex items-start space-x-3">
                <div className="flex-1">
                  <div className="flex items-center space-x-2 mb-1">
                    <span className="text-lg">{log.step}</span>
                    <span className="text-xs text-gray-500">
                      {formatTimestamp(log.timestamp)}
                    </span>
                  </div>
                  <p className="text-sm text-gray-700">{log.message}</p>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};

export default WebSocketLogs;
