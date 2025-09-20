'use client'

import { useState, useEffect } from 'react'
import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { 
  LayoutDashboard, 
  Car, 
  Database, 
  Upload, 
  Package, 
  User, 
  Settings,
  Menu,
  X,
  Users
} from 'lucide-react'

interface User {
  role: 'admin' | 'operacao' | 'atendimento';
}

const navigation = [
  { 
    name: 'Dashboard', 
    href: '/dashboard', 
    icon: LayoutDashboard,
    roles: ['admin', 'operacao', 'atendimento']
  },
  { 
    name: 'Car Plate', 
    href: '/dashboard/car-plate', 
    icon: Car,
    roles: ['admin', 'operacao', 'atendimento']
  },
  { 
    name: 'Stock', 
    href: '/dashboard/stock', 
    icon: Package,
    roles: ['admin', 'operacao', 'atendimento']
  },
  { 
    name: 'DePara', 
    href: '/dashboard/depara', 
    icon: Database,
    roles: ['admin', 'operacao', 'atendimento']
  },
  { 
    name: 'Import XML', 
    href: '/dashboard/import-xml', 
    icon: Upload,
    roles: ['admin', 'operacao']
  },
  { 
    name: 'Integration', 
    href: '/dashboard/integration', 
    icon: Settings,
    roles: ['admin', 'operacao']
  },
  { 
    name: 'XML Integrator', 
    href: '/dashboard/xml-integrator', 
    icon: Upload,
    roles: ['admin', 'operacao']
  },
  { 
    name: 'Usuários', 
    href: '/dashboard/users', 
    icon: Users,
    roles: ['admin']
  },
]

const profileNavigation = [
  { name: 'Perfil', href: '/dashboard/profile', icon: User },
  { name: 'Configurações', href: '/dashboard/settings', icon: Settings, roles: ['admin'] },
]

export default function Sidebar() {
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false)
  const [user, setUser] = useState<User | null>(null)
  const pathname = usePathname()

  const isActive = (href: string) => pathname === href

  useEffect(() => {
    // Get user role from token or API
    const token = localStorage.getItem('token')
    if (token) {
      try {
        const payload = JSON.parse(atob(token.split('.')[1]))
        setUser({ role: payload.role || 'atendimento' })
      } catch (error) {
        console.error('Error parsing token:', error)
        setUser({ role: 'atendimento' })
      }
    }
  }, [])

  const hasPermission = (requiredRoles: string[]) => {
    if (!user) return false
    return requiredRoles.includes(user.role)
  }

  const filteredNavigation = navigation.filter(item => hasPermission(item.roles))

  return (
    <>
      {/* Mobile menu button */}
      <div className="lg:hidden fixed top-4 left-4 z-50">
        <button
          onClick={() => setIsMobileMenuOpen(!isMobileMenuOpen)}
          className="p-2 rounded-lg bg-white shadow-md border border-gray-200"
        >
          {isMobileMenuOpen ? <X className="w-5 h-5" /> : <Menu className="w-5 h-5" />}
        </button>
      </div>

      {/* Mobile overlay */}
      {isMobileMenuOpen && (
        <div 
          className="lg:hidden fixed inset-0 bg-black bg-opacity-50 z-40"
          onClick={() => setIsMobileMenuOpen(false)}
        />
      )}

      {/* Sidebar */}
      <div className={`
        fixed inset-y-0 left-0 z-50 w-64 bg-white shadow-lg transform transition-transform duration-300 ease-in-out lg:translate-x-0 lg:static lg:inset-0
        ${isMobileMenuOpen ? 'translate-x-0' : '-translate-x-full'}
      `}>
        <div className="flex flex-col h-full">
          {/* Logo */}
          <div className="flex items-center px-6 py-4 border-b border-gray-200">
            <div className="w-8 h-8 bg-gradient-to-r from-blue-600 to-indigo-600 rounded-lg flex items-center justify-center">
              <Package className="w-5 h-5 text-white" />
            </div>
            <div className="ml-3">
              <h1 className="text-lg font-bold text-gray-900">AMZ Tools</h1>
              <p className="text-xs text-gray-500">Portal</p>
            </div>
          </div>

          {/* Navigation */}
          <nav className="flex-1 px-4 py-6 space-y-2 overflow-y-auto">
            <div className="space-y-1">
              <h3 className="px-3 text-xs font-semibold text-gray-500 uppercase tracking-wider">
                Módulos
              </h3>
              {filteredNavigation.map((item) => {
                const Icon = item.icon
                return (
                  <Link
                    key={item.name}
                    href={item.href}
                    onClick={() => setIsMobileMenuOpen(false)}
                    className={`
                      group flex items-center px-3 py-2 text-sm font-medium rounded-lg transition-colors
                      ${isActive(item.href)
                        ? 'bg-blue-50 text-blue-700 border-r-2 border-blue-700'
                        : 'text-gray-700 hover:bg-gray-50 hover:text-gray-900'
                      }
                    `}
                  >
                    <Icon className={`
                      mr-3 h-5 w-5 flex-shrink-0
                      ${isActive(item.href) ? 'text-blue-500' : 'text-gray-400 group-hover:text-gray-500'}
                    `} />
                    {item.name}
                  </Link>
                )
              })}
            </div>

            <div className="pt-6 space-y-1">
              <h3 className="px-3 text-xs font-semibold text-gray-500 uppercase tracking-wider">
                Conta
              </h3>
              {profileNavigation.filter(item => !item.roles || hasPermission(item.roles)).map((item) => {
                const Icon = item.icon
                return (
                  <Link
                    key={item.name}
                    href={item.href}
                    onClick={() => setIsMobileMenuOpen(false)}
                    className={`
                      group flex items-center px-3 py-2 text-sm font-medium rounded-lg transition-colors
                      ${isActive(item.href)
                        ? 'bg-blue-50 text-blue-700 border-r-2 border-blue-700'
                        : 'text-gray-700 hover:bg-gray-50 hover:text-gray-900'
                      }
                    `}
                  >
                    <Icon className={`
                      mr-3 h-5 w-5 flex-shrink-0
                      ${isActive(item.href) ? 'text-blue-500' : 'text-gray-400 group-hover:text-gray-500'}
                    `} />
                    {item.name}
                  </Link>
                )
              })}
            </div>
          </nav>

          {/* Footer */}
          <div className="p-4 border-t border-gray-200">
            <div className="text-xs text-gray-500 text-center">
              <p>AMZ Web Tools Portal</p>
              <p className="mt-1">v1.0.0</p>
            </div>
          </div>
        </div>
      </div>
    </>
  )
}

