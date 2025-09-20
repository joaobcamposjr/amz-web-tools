'use client'

import { useState, useEffect } from 'react'
import { Car, Search, Clock, CheckCircle, AlertCircle, Database, Globe } from 'lucide-react'
import JsonRenderer from '@/components/JsonRenderer'

export default function CarPlatePage() {
  const [plate, setPlate] = useState('')
  const [loading, setLoading] = useState(false)
  const [result, setResult] = useState<any>(null)
  const [error, setError] = useState('')
  const [source, setSource] = useState<'cache' | 'api' | null>(null)
  const [chassi, setChassi] = useState<string | null>(null)
  const [brandLogo, setBrandLogo] = useState<string | null>(null)
  const [searchHistory, setSearchHistory] = useState<any[]>([])
  const [historyLoading, setHistoryLoading] = useState(false)

  // Load search history on component mount
  useEffect(() => {
    loadSearchHistory()
  }, [])

  const loadSearchHistory = async () => {
    setHistoryLoading(true)
    try {
      const response = await fetch('/api/v1/car-plate/history', {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
        },
      })
      const data = await response.json()
      if (data.success) {
        setSearchHistory(data.data.history || [])
      }
    } catch (error) {
      console.error('Error loading search history:', error)
    } finally {
      setHistoryLoading(false)
    }
  }

  const handleSearch = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!plate.trim()) return

    setLoading(true)
    setError('')
    setResult(null)

    try {
      const response = await fetch(`/api/v1/car-plate/${plate.toUpperCase()}`, {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
        },
      })

      const data = await response.json()

      if (data.success) {
        setResult(data.data.plate_data)
        setSource(data.data.source) // 'cache' or 'api'
        setChassi(data.data.chassi || null) // Extract chassi if available
        setBrandLogo(data.data.brand_logo || null) // Extract brand logo if available
        setError('')
        // Reload search history to get real data
        loadSearchHistory()
      } else {
        setError(data.message || 'Erro ao consultar placa')
        setResult(null)
        setSource(null)
        setChassi(null)
        setBrandLogo(null)
        // Reload search history to get real data
        loadSearchHistory()
      }
    } catch (error) {
      setError('Erro de conexão. Verifique se o backend está rodando.')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="bg-white rounded-xl p-6 shadow-sm border border-gray-100">
        <div className="flex items-center space-x-3">
          <div className="p-2 bg-blue-50 rounded-lg">
            <Car className="w-6 h-6 text-blue-600" />
          </div>
          <div>
            <h1 className="text-2xl font-bold text-gray-900">Consulta de Placa</h1>
            <p className="text-gray-600">Busque informações detalhadas sobre veículos</p>
          </div>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Search Form */}
        <div className="lg:col-span-2">
          <div className="bg-white rounded-xl p-6 shadow-sm border border-gray-100">
            <h2 className="text-lg font-semibold text-gray-900 mb-4">Nova Consulta</h2>
            
            <form onSubmit={handleSearch} className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Placa do Veículo
                </label>
                <div className="relative">
                  <input
                    type="text"
                    value={plate}
                    onChange={(e) => setPlate(e.target.value.toUpperCase())}
                    placeholder="ABC1234"
                    maxLength={7}
                    className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent text-lg font-mono tracking-wider"
                  />
                  <Car className="absolute right-3 top-1/2 transform -translate-y-1/2 text-gray-400 w-5 h-5" />
                </div>
                <p className="text-sm text-gray-500 mt-1">
                  Digite a placa sem hífen (ex: ABC1234)
                </p>
              </div>

              <button
                type="submit"
                disabled={loading || !plate.trim()}
                className="w-full bg-gradient-to-r from-blue-600 to-indigo-600 text-white py-3 px-4 rounded-lg font-medium hover:from-blue-700 hover:to-indigo-700 focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 transition-all disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {loading ? (
                  <div className="flex items-center justify-center">
                    <div className="w-5 h-5 border-2 border-white border-t-transparent rounded-full animate-spin mr-2"></div>
                    Consultando...
                  </div>
                ) : (
                  <div className="flex items-center justify-center">
                    <Search className="w-5 h-5 mr-2" />
                    Consultar Placa
                  </div>
                )}
              </button>

              {/* Sample Plates */}
              <div className="mt-4">
                <p className="text-sm text-gray-600 mb-2">Placas de exemplo:</p>
                <div className="flex flex-wrap gap-2">
                  {['KRF3901', 'ABC1234', 'XYZ9876'].map((samplePlate) => (
                    <button
                      key={samplePlate}
                      onClick={() => setPlate(samplePlate)}
                      className="px-3 py-1 text-xs bg-gray-100 hover:bg-gray-200 rounded-full font-mono transition-colors"
                    >
                      {samplePlate}
                    </button>
                  ))}
                </div>
              </div>
            </form>

            {/* Error Message */}
            {error && (
              <div className="mt-4 p-4 bg-red-50 border border-red-200 rounded-lg">
                <div className="flex items-center">
                  <AlertCircle className="w-5 h-5 text-red-600 mr-2" />
                  <span className="text-red-800">{error}</span>
                </div>
              </div>
            )}

            {/* Result */}
            {result && (
              <div className="mt-6">
                <div className="flex items-center mb-4 p-4 bg-green-50 border border-green-200 rounded-lg">
                  <CheckCircle className="w-5 h-5 text-green-600 mr-2" />
                  <span className="text-green-800 font-medium">Consulta realizada com sucesso!</span>
                  <div className="ml-auto flex items-center space-x-2">
                    {source === 'cache' ? (
                      <>
                        <Database className="w-4 h-4 text-blue-600" />
                        <span className="text-sm text-blue-600">Cache</span>
                      </>
                    ) : (
                      <>
                        <Globe className="w-4 h-4 text-orange-600" />
                        <span className="text-sm text-orange-600">API</span>
                      </>
                    )}
                  </div>
                </div>
                
                {/* Brand Logo */}
                {brandLogo && (
                  <div className="mb-4 p-4 bg-white border border-gray-200 rounded-lg shadow-sm">
                    <div className="flex items-center justify-center">
                      <img 
                        src={brandLogo} 
                        alt="Logo da marca" 
                        className="h-12 w-auto object-contain"
                        onError={(e) => {
                          e.currentTarget.style.display = 'none'
                        }}
                      />
                    </div>
                  </div>
                )}
                
                {/* Chassi Information */}
                {chassi && (
                  <div className="mb-4 p-3 bg-blue-50 border border-blue-200 rounded-lg">
                    <div className="flex items-center">
                      <Car className="w-4 h-4 text-blue-600 mr-2" />
                      <span className="text-sm font-medium text-blue-800">Chassi:</span>
                      <span className="ml-2 font-mono text-sm text-blue-900">{chassi}</span>
                    </div>
                  </div>
                )}
                
                <JsonRenderer 
                  data={result} 
                  title={`Dados da Placa ${plate.toUpperCase()}`}
                />
              </div>
            )}
          </div>
        </div>

        {/* Search History */}
        <div>
          <div className="bg-white rounded-xl p-6 shadow-sm border border-gray-100">
            <h2 className="text-lg font-semibold text-gray-900 mb-4">Histórico de Consultas</h2>
            
            <div className="space-y-3">
              {historyLoading ? (
                <div className="flex items-center justify-center py-4">
                  <div className="w-5 h-5 border-2 border-blue-600 border-t-transparent rounded-full animate-spin"></div>
                  <span className="ml-2 text-sm text-gray-600">Carregando histórico...</span>
                </div>
              ) : searchHistory.length === 0 ? (
                <div className="text-center py-4 text-gray-500 text-sm">
                  Nenhuma consulta realizada ainda
                </div>
              ) : Array.isArray(searchHistory) ? (
                searchHistory.map((item, index) => (
                  <div key={`${item.plate}-${item.created_at}-${index}`} className="flex items-center space-x-3 p-3 bg-gray-50 rounded-lg">
                    <div className={`p-1 rounded-full ${
                      item.status === 'success' ? 'bg-green-100' : 'bg-red-100'
                    }`}>
                      {item.status === 'success' ? (
                        <CheckCircle className="w-4 h-4 text-green-600" />
                      ) : (
                        <AlertCircle className="w-4 h-4 text-red-600" />
                      )}
                    </div>
                    <div className="flex-1">
                      <p className="font-mono text-sm font-medium">{item.plate}</p>
                      <p className="text-xs text-gray-500">
                        {new Date(item.created_at).toLocaleString('pt-BR')}
                      </p>
                    </div>
                    <Clock className="w-4 h-4 text-gray-400" />
                  </div>
                ))
              ) : (
                <div className="text-center py-4 text-gray-500 text-sm">
                  Erro ao carregar histórico
                </div>
              )}
            </div>

            <button className="w-full mt-4 text-sm text-blue-600 hover:text-blue-700 font-medium">
              Ver histórico completo
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}
