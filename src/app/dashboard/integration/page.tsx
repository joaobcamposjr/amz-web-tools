'use client';

import React, { useState } from 'react';
import WebSocketLogs from '@/components/WebSocketLogs';

interface IntegrationRequest {
  conta: string;
  marketplace: string;
  num_pedido: string;
}

interface IntegrationResponse {
  total_processed: number;
  success_count: number;
  error_count: number;
  results: Array<{
    pedido: string;
    numero_pedido: string;
    status: string;
  }>;
  logs: Array<{
    timestamp: string;
    level: string;
    step: string;
    message: string;
  }>;
}

export default function IntegrationPage() {
  const [formData, setFormData] = useState<IntegrationRequest>({
    conta: '',
    marketplace: '',
    num_pedido: '',
  });
  const [isProcessing, setIsProcessing] = useState(false);
  const [result, setResult] = useState<IntegrationResponse | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [currentProcessId, setCurrentProcessId] = useState<string | null>(null);

  const contaOptions = [
    { value: 'principal', label: 'Principal' },
    { value: 'oficial', label: 'Oficial' },
    { value: 'psa', label: 'PSA' },
    { value: 'jeep', label: 'Jeep' },
    { value: 'renault', label: 'Renault' },
    { value: 'ford', label: 'Ford' },
  ];

  const marketplaceOptions = [
    { value: 'mercadolivre', label: 'Mercado Livre' },
  ];

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => {
    const { name, value } = e.target;
    setFormData(prev => ({
      ...prev,
      [name]: value
    }));
  };

  const handleProcess = async () => {
    if (!formData.conta || !formData.marketplace || !formData.num_pedido) {
      setError('Todos os campos são obrigatórios');
      return;
    }

    setIsProcessing(true);
    setError(null);
    setResult(null);
    setCurrentProcessId(`integration_${Date.now()}`);

    try {
      const response = await fetch('/api/v1/integration/execute', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
        },
        body: JSON.stringify(formData),
      });

      const data = await response.json();

      if (response.ok) {
        setResult(data.data);
        console.log('🔍 API Response data:', data);
        console.log('🔍 API Response data.data:', data.data);
        console.log('🔍 API Response data.data.logs:', data.data?.logs);
      } else {
        setError(data.message || 'Erro ao processar integração');
      }
    } catch (err) {
      console.error('Erro ao processar integração:', err);
      setError('Erro de conexão com o servidor');
    } finally {
      setIsProcessing(false);
    }
  };

  const handleLogReceived = (log: any) => {
    console.log('Log recebido:', log);
  };

  return (
    <div className="min-h-screen bg-gray-50 p-6">
      <div className="max-w-4xl mx-auto">
        <div className="bg-white rounded-lg shadow-md p-6 mb-6">
          <h1 className="text-2xl font-bold text-gray-800 mb-6">Integração de Pedidos</h1>
          
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-6">
            {/* Conta */}
            <div>
              <label htmlFor="conta" className="block text-sm font-medium text-gray-700 mb-2">
                Conta *
              </label>
              <select
                id="conta"
                name="conta"
                value={formData.conta}
                onChange={handleInputChange}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                disabled={isProcessing}
              >
                <option value="">Selecione uma conta</option>
                {contaOptions.map(option => (
                  <option key={option.value} value={option.value}>
                    {option.label}
                  </option>
                ))}
              </select>
            </div>

            {/* Marketplace */}
            <div>
              <label htmlFor="marketplace" className="block text-sm font-medium text-gray-700 mb-2">
                Marketplace *
              </label>
              <select
                id="marketplace"
                name="marketplace"
                value={formData.marketplace}
                onChange={handleInputChange}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                disabled={isProcessing}
              >
                <option value="">Selecione um marketplace</option>
                {marketplaceOptions.map(option => (
                  <option key={option.value} value={option.value}>
                    {option.label}
                  </option>
                ))}
              </select>
            </div>

            {/* Número do Pedido */}
            <div>
              <label htmlFor="num_pedido" className="block text-sm font-medium text-gray-700 mb-2">
                Número do Pedido *
              </label>
              <input
                type="text"
                id="num_pedido"
                name="num_pedido"
                value={formData.num_pedido}
                onChange={handleInputChange}
                placeholder="Ex: 2000008982893407"
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                disabled={isProcessing}
              />
            </div>
          </div>

          {/* Botão de Processar */}
          <div className="flex justify-center">
            <button
              onClick={handleProcess}
              disabled={isProcessing || !formData.conta || !formData.marketplace || !formData.num_pedido}
              className={`px-8 py-3 rounded-md font-medium transition-colors ${
                isProcessing || !formData.conta || !formData.marketplace || !formData.num_pedido
                  ? 'bg-gray-400 text-gray-600 cursor-not-allowed'
                  : 'bg-blue-600 text-white hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2'
              }`}
            >
              {isProcessing ? (
                <div className="flex items-center">
                  <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2"></div>
                  Processando...
                </div>
              ) : (
                'Iniciar Integração'
              )}
            </button>
          </div>

          {/* Mensagem de Erro */}
          {error && (
            <div className="mt-4 p-4 bg-red-100 border border-red-400 text-red-700 rounded-md">
              {error}
            </div>
          )}
        </div>

        {/* Logs em Tempo Real */}
        {currentProcessId && (
          <div className="bg-white rounded-lg shadow-md p-6">
            <h2 className="text-xl font-bold text-gray-800 mb-4">Logs de Processamento</h2>
            <WebSocketLogs
              processId={currentProcessId}
              onLogReceived={handleLogReceived}
              processResult={result}
            />
          </div>
        )}

      </div>
    </div>
  );
}
