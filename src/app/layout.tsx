import type { Metadata } from 'next'
import './globals.css'

export const metadata: Metadata = {
  title: 'AMZ Web Tools Portal',
  description: 'Portal de autopeças com sistema de login e módulos específicos',
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="pt-BR">
      <body className="bg-gray-50 antialiased font-sans">
        {children}
      </body>
    </html>
  )
}

