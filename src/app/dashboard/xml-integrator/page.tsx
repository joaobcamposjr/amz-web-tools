'use client';

import { useState } from 'react';
import { Play, CheckCircle, XCircle, Clock, AlertCircle } from 'lucide-react';

interface XMLIntegrationResult {
  total_processed: number;
  success_count: number;
  error_count: number;
  results: Array<{
    pedido: string;
    prenota: string;
    envio: string;
    nota_fiscal: string;
    status: string;
    substatus: string;
  }>;
}

export default function XMLIntegratorPage() {
  const [numPedido, setNumPedido] = useState('');
  const [isProcessing, setIsProcessing] = useState(false);
  const [result, setResult] = useState<XMLIntegrationResult | null>(null);
  const [error, setError] = useState<string | null>(null);

  const handleProcess = async () => {
    if (!numPedido.trim()) {
      setError('Por favor, informe o número do pedido');
      return;
    }

    setIsProcessing(true);
    setError(null);
    setResult(null);

    try {
      const response = await fetch('/api/v1/xml-integrator/process', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
        },
        body: JSON.stringify({
          num_pedido: numPedido.trim(),
        }),
      });

      const data = await response.json();

      if (response.ok) {
        setResult(data.data);
      } else {
        setError(data.message || 'Erro ao processar integração XML');
      }
    } catch (err) {
      setError('Erro de conexão. Verifique sua internet e tente novamente.');
    } finally {
      setIsProcessing(false);
    }
  };

  const getStatusIcon = (status: string, substatus: string) => {
    if (status === 'ready_to_ship' && substatus === 'invoice_pending') {
      return <Clock className="w-5 h-5 text-blue-500" />;
    } else if (status === 'pending' && substatus === 'buffered') {
      return <AlertCircle className="w-5 h-5 text-yellow-500" />;
    } else {
      return <CheckCircle className="w-5 h-5 text-green-500" />;
    }
  };

  const getStatusText = (status: string, substatus: string) => {
    if (status === 'ready_to_ship' && substatus === 'invoice_pending') {
      return 'Pronto para envio - XML pendente';
    } else if (status === 'pending' && substatus === 'buffered') {
      return 'Aguardando agendamento';
    } else {
      return 'Envio Flex';
    }
  };

  const getStatusColor = (status: string, substatus: string) => {
    if (status === 'ready_to_ship' && substatus === 'invoice_pending') {
      return 'bg-blue-100 text-blue-800';
    } else if (status === 'pending' && substatus === 'buffered') {
      return 'bg-yellow-100 text-yellow-800';
    } else {
      return 'bg-green-100 text-green-800';
    }
  };

  return (
    <div className="min-h-screen bg-gray-50 p-6">
      <div className="max-w-7xl mx-auto">
        {/* Header */}
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-gray-900 mb-2">XML Integrator</h1>
          <p className="text-gray-600">
            Processe integração de XMLs com o Mercado Livre para pedidos específicos
          </p>
        </div>

        {/* Input Section */}
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6 mb-8">
          <h2 className="text-xl font-semibold text-gray-900 mb-4">Processar Pedido</h2>
          
          <div className="flex items-center space-x-4">
            <div className="flex-grow">
              <label htmlFor="numPedido" className="block text-sm font-medium text-gray-700 mb-2">
                Número do Pedido
              </label>
              <input
                id="numPedido"
                type="text"
                placeholder="Digite o número do pedido (ex: 1234567890)"
                value={numPedido}
                onChange={(e) => setNumPedido(e.target.value)}
                onKeyPress={(e) => e.key === 'Enter' && handleProcess()}
                className="w-full p-3 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all duration-200"
                disabled={isProcessing}
              />
            </div>
            
            <div className="pt-6">
              <button
                onClick={handleProcess}
                disabled={isProcessing || !numPedido.trim()}
                className="bg-blue-600 text-white px-6 py-3 rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 transition-colors duration-200 disabled:opacity-50 disabled:cursor-not-allowed flex items-center space-x-2"
              >
                {isProcessing ? (
                  <>
                    <Clock className="w-5 h-5 animate-spin" />
                    <span>Processando...</span>
                  </>
                ) : (
                  <>
                    <Play className="w-5 h-5" />
                    <span>Processar</span>
                  </>
                )}
              </button>
            </div>
          </div>
        </div>

        {/* Error Display */}
        {error && (
          <div className="bg-red-50 border border-red-200 rounded-lg p-4 mb-8">
            <div className="flex items-center">
              <XCircle className="w-5 h-5 text-red-500 mr-2" />
              <span className="text-red-700 font-medium">Erro:</span>
            </div>
            <p className="text-red-600 mt-1">{error}</p>
          </div>
        )}

        {/* Results Display */}
        {result && (
          <div className="space-y-6">
            {/* Summary */}
            <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
              <h2 className="text-xl font-semibold text-gray-900 mb-4">Resumo da Execução</h2>
              
              <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
                <div className="bg-gray-50 rounded-lg p-4 text-center">
                  <div className="text-2xl font-bold text-gray-900">{result.total_processed}</div>
                  <div className="text-sm text-gray-600">Total Processados</div>
                </div>
                
                <div className="bg-green-50 rounded-lg p-4 text-center">
                  <div className="text-2xl font-bold text-green-600">{result.success_count}</div>
                  <div className="text-sm text-green-600">Sucessos</div>
                </div>
                
                <div className="bg-red-50 rounded-lg p-4 text-center">
                  <div className="text-2xl font-bold text-red-600">{result.error_count}</div>
                  <div className="text-sm text-red-600">Erros</div>
                </div>
                
                <div className="bg-blue-50 rounded-lg p-4 text-center">
                  <div className="text-2xl font-bold text-blue-600">
                    {result.total_processed > 0 ? Math.round((result.success_count / result.total_processed) * 100) : 0}%
                  </div>
                  <div className="text-sm text-blue-600">Taxa de Sucesso</div>
                </div>
              </div>
            </div>

            {/* Detailed Results */}
            {result.results.length > 0 && (
              <div className="bg-white rounded-lg shadow-sm border border-gray-200">
                <div className="p-6 border-b border-gray-200">
                  <h2 className="text-xl font-semibold text-gray-900">Detalhes dos Resultados</h2>
                </div>
                
                <div className="overflow-x-auto">
                  <table className="min-w-full divide-y divide-gray-200">
                    <thead className="bg-gray-50">
                      <tr>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                          Pedido
                        </th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                          Prenota
                        </th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                          Envio
                        </th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                          Nota Fiscal
                        </th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                          Status ML
                        </th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                          Substatus
                        </th>
                      </tr>
                    </thead>
                    <tbody className="bg-white divide-y divide-gray-200">
                      {result.results.map((item, index) => (
                        <tr key={index} className="hover:bg-gray-50">
                          <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                            {item.pedido}
                          </td>
                          <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                            {item.prenota}
                          </td>
                          <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                            {item.envio}
                          </td>
                          <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                            {item.nota_fiscal}
                          </td>
                          <td className="px-6 py-4 whitespace-nowrap">
                            <div className="flex items-center">
                              {getStatusIcon(item.status, item.substatus)}
                              <span className={`ml-2 px-2 py-1 rounded-full text-xs font-medium ${getStatusColor(item.status, item.substatus)}`}>
                                {getStatusText(item.status, item.substatus)}
                              </span>
                            </div>
                          </td>
                          <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                            {item.substatus}
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              </div>
            )}
          </div>
        )}

        {/* Info Section */}
        <div className="bg-blue-50 border border-blue-200 rounded-lg p-6 mt-8">
          <h3 className="text-lg font-semibold text-blue-900 mb-2">ℹ️ Como funciona</h3>
          <ul className="text-blue-800 space-y-2">
            <li>• Digite o número do pedido que deseja processar</li>
            <li>• O sistema buscará os dados do pedido no PostgreSQL</li>
            <li>• Verificará o XML correspondente no Oracle NBS</li>
            <li>• Consultará o status do envio no Mercado Livre</li>
            <li>• Enviará o XML automaticamente se necessário</li>
            <li>• Atualizará os status nos bancos de dados</li>
            <li>• Enviará notificações via Telegram sobre o progresso</li>
          </ul>
        </div>
      </div>
    </div>
  );
}




