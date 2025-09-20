'use client'

import { useState } from 'react'
import { ChevronDown, ChevronRight, Copy, Check } from 'lucide-react'

interface JsonRendererProps {
  data: any
  title?: string
  level?: number
}

export default function JsonRenderer({ data, title, level = 0 }: JsonRendererProps) {
  const [isExpanded, setIsExpanded] = useState(level < 2) // Auto-expand first 2 levels
  const [copied, setCopied] = useState(false)

  const copyToClipboard = async (text: string) => {
    try {
      await navigator.clipboard.writeText(text)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    } catch (err) {
      console.error('Failed to copy:', err)
    }
  }

  const renderValue = (value: any, key: string): React.ReactNode => {
    if (value === null) {
      return <span className="text-gray-500 italic">null</span>
    }

    if (typeof value === 'string') {
      return (
        <span className="text-green-700 font-mono">
          "{value}"
        </span>
      )
    }

    if (typeof value === 'number') {
      return <span className="text-blue-600 font-mono">{value}</span>
    }

    if (typeof value === 'boolean') {
      return (
        <span className={`font-mono ${value ? 'text-green-600' : 'text-red-600'}`}>
          {value.toString()}
        </span>
      )
    }

    if (Array.isArray(value)) {
      return (
        <div className="ml-4">
          <span className="text-gray-600">[</span>
          <div className="ml-4">
            {value.map((item, index) => (
              <div key={index} className="flex items-start">
                <span className="text-gray-500 mr-2">{index}:</span>
                <div className="flex-1">
                  {renderValue(item, `${key}[${index}]`)}
                </div>
              </div>
            ))}
          </div>
          <span className="text-gray-600">]</span>
        </div>
      )
    }

    if (typeof value === 'object') {
      return (
        <div className="ml-4">
          <span className="text-gray-600">{'{'}</span>
          <div className="ml-4">
            {Object.entries(value).map(([objKey, objValue]) => (
              <div key={objKey} className="flex items-start">
                <span className="text-purple-600 font-medium mr-2">
                  "{objKey}":
                </span>
                <div className="flex-1">
                  {renderValue(objValue, `${key}.${objKey}`)}
                </div>
              </div>
            ))}
          </div>
          <span className="text-gray-600">{'}'}</span>
        </div>
      )
    }

    return <span className="text-gray-500">{String(value)}</span>
  }

  if (typeof data !== 'object' || data === null) {
    return (
      <div className="bg-white rounded-lg p-4 border border-gray-200">
        {title && (
          <h3 className="text-lg font-semibold text-gray-900 mb-2">{title}</h3>
        )}
        <div className="font-mono text-sm">
          {renderValue(data, 'root')}
        </div>
      </div>
    )
  }

  const isObject = !Array.isArray(data)
  const entries = isObject ? Object.entries(data) : data.map((item, index) => [index, item])

  return (
    <div className="bg-white rounded-lg border border-gray-200 overflow-hidden">
      {title && (
        <div className="bg-gray-50 px-4 py-3 border-b border-gray-200 flex items-center justify-between">
          <h3 className="text-lg font-semibold text-gray-900">{title}</h3>
          <button
            onClick={() => copyToClipboard(JSON.stringify(data, null, 2))}
            className="flex items-center space-x-1 text-sm text-gray-600 hover:text-gray-800"
          >
            {copied ? (
              <>
                <Check className="w-4 h-4 text-green-600" />
                <span className="text-green-600">Copiado!</span>
              </>
            ) : (
              <>
                <Copy className="w-4 h-4" />
                <span>Copiar JSON</span>
              </>
            )}
          </button>
        </div>
      )}

      <div className="p-4">
        <div className="font-mono text-sm">
          <span className="text-gray-600">{isObject ? '{' : '['}</span>
          
          {isExpanded ? (
            <div className="ml-4 mt-2">
              {entries.map(([key, value], index) => (
                <div key={key} className="mb-2">
                  <div className="flex items-start">
                    <span className="text-purple-600 font-medium mr-2">
                      "{key}":
                    </span>
                    <div className="flex-1">
                      {renderValue(value, String(key))}
                    </div>
                  </div>
                  {index < entries.length - 1 && <span className="text-gray-400">,</span>}
                </div>
              ))}
            </div>
          ) : (
            <span className="text-gray-500 ml-2">
              {isObject ? `${entries.length} propriedades` : `${entries.length} itens`}...
            </span>
          )}
          
          <span className="text-gray-600">{isObject ? '}' : ']'}</span>
        </div>

        {entries.length > 0 && (
          <button
            onClick={() => setIsExpanded(!isExpanded)}
            className="mt-3 flex items-center space-x-1 text-sm text-blue-600 hover:text-blue-800"
          >
            {isExpanded ? (
              <>
                <ChevronDown className="w-4 h-4" />
                <span>Recolher</span>
              </>
            ) : (
              <>
                <ChevronRight className="w-4 h-4" />
                <span>Expandir</span>
              </>
            )}
          </button>
        )}
      </div>
    </div>
  )
}

