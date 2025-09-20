'use client';

import React, { useState, useEffect } from 'react';
import { Search, Plus, Eye, Edit, Trash2, Check, X } from 'lucide-react';
import ImageSlider from '@/components/ImageSlider';
import axios from 'axios';

interface DeParaProduct {
  id: string;
  mlbu: string;
  type: string;
  sku: string;
  company: string;
  permalink: string;
  ship_cost_slow: number;
  ship_cost_standard: number;
  ship_cost_nextday: number;
  pictures: string[];
  updated_at: string;
  created_at: string;
}

interface IntegrationTable {
  id: string;
  table_name: string;
  display_name: string;
  is_active: boolean;
  created_at: string;
}

interface TableOptions {
  empresa: string[];
  conta: string[];
  marketplace: string[];
}

export default function DeParaPage() {
  const [tables, setTables] = useState<IntegrationTable[]>([]);
  const [tableOptions, setTableOptions] = useState<TableOptions>({ empresa: [], conta: [], marketplace: [] });
  const [selectedEmpresa, setSelectedEmpresa] = useState('amazonas');
  const [selectedConta, setSelectedConta] = useState('psa');
  const [selectedMarketplace, setSelectedMarketplace] = useState('mercadolivre');
  const [searchQuery, setSearchQuery] = useState('');
  const [products, setProducts] = useState<DeParaProduct[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [currentPage, setCurrentPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [totalCount, setTotalCount] = useState(0);
  const [pageSize] = useState(15); // 3 linhas × 5 colunas = 15 cards
  const [allProducts, setAllProducts] = useState<DeParaProduct[]>([]);
  const [lastSearchQuery, setLastSearchQuery] = useState('');
  
  // Modal states
  const [selectedProduct, setSelectedProduct] = useState<DeParaProduct | null>(null);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [isEditing, setIsEditing] = useState(false);
  const [editForm, setEditForm] = useState({ sku: '', company: '' });
  
  // Create form states
  const [showCreateForm, setShowCreateForm] = useState(false);
  const [createForm, setCreateForm] = useState({ id: '', sku: '', company: '', mlbu: '', type: '' });

  // Load available tables and options on component mount
  useEffect(() => {
    loadTables();
    loadTableOptions();
  }, []);

  const loadTables = async () => {
    try {
      const response = await axios.get('/api/v1/test/depara/tables');
      if (response.data.success) {
        setTables(response.data.data);
      }
    } catch (err) {
      console.error('Error loading tables:', err);
    }
  };

  const loadTableOptions = async () => {
    try {
      const response = await axios.get('/api/v1/test/depara/options');
      if (response.data.success) {
        setTableOptions(response.data.data);
      }
    } catch (err) {
      console.error('Error loading table options:', err);
      // Fallback to hardcoded options
      const options = {
        empresa: ['amazonas'],
        conta: ['psa'],
        marketplace: ['mercadolivre']
      };
      setTableOptions(options);
    }
  };

  const getCurrentTableName = () => {
    return `integration.${selectedEmpresa}_${selectedConta}.${selectedMarketplace}_base`;
  };

  const detectSearchType = (query: string): string => {
    const trimmedQuery = query.trim();
    if (trimmedQuery.startsWith('MLBU')) {
      return 'mlbu';
    } else if (trimmedQuery.startsWith('MLB') && trimmedQuery.length >= 10) {
      return 'id';
    } else {
      return 'sku';
    }
  };

  const handleSearch = async (page: number = 1) => {
    if (!searchQuery.trim()) return;
    
    // If we already have the products loaded and it's just a page change, just paginate
    if (allProducts.length > 0 && page !== 1 && searchQuery === lastSearchQuery) {
      paginateResults(page);
      return;
    }
    
    setLoading(true);
    setError('');
    
    try {
      const searchType = detectSearchType(searchQuery);
      const response = await axios.post(`/api/v1/test/depara/search`, {
        table_name: getCurrentTableName(),
        query: searchQuery.trim(),
        search_by: searchType
      });
      
      if (response.data.success) {
        const allProductsData = response.data.data.products || [];
        setAllProducts(allProductsData);
        setTotalCount(allProductsData.length);
        setTotalPages(Math.ceil(allProductsData.length / pageSize));
        setCurrentPage(page);
        setLastSearchQuery(searchQuery.trim());
        
        // Paginate the results
        paginateResults(page, allProductsData);
      } else {
        setError(response.data.message);
        setProducts([]);
        setAllProducts([]);
        setTotalCount(0);
        setTotalPages(1);
        setCurrentPage(1);
      }
    } catch (err: any) {
      setError(err.response?.data?.message || 'Erro ao pesquisar produtos');
      setProducts([]);
      setAllProducts([]);
      setTotalCount(0);
      setTotalPages(1);
      setCurrentPage(1);
    } finally {
      setLoading(false);
    }
  };

  const paginateResults = (page: number, productsData?: DeParaProduct[]) => {
    const dataToUse = productsData || allProducts;
    const startIndex = (page - 1) * pageSize;
    const endIndex = startIndex + pageSize;
    const paginatedProducts = dataToUse.slice(startIndex, endIndex);
    
    setProducts(paginatedProducts);
    setCurrentPage(page);
  };

  const handleProductClick = (product: DeParaProduct) => {
    setSelectedProduct(product);
    setEditForm({ sku: product.sku, company: product.company });
    setIsEditing(false);
    setIsModalOpen(true);
  };

  const handleEdit = () => {
    setIsEditing(true);
  };

  const handleUpdate = async () => {
    if (!selectedProduct) return;
    
    // Validação dos campos obrigatórios
    if (!editForm.sku.trim() || !editForm.company.trim()) {
      alert('SKU e Empresa são campos obrigatórios.');
      return;
    }
    
    // Confirmação antes de atualizar
    if (!confirm('Tem certeza que deseja salvar as alterações deste produto?')) {
      return;
    }
    
    try {
      const response = await axios.put(`/api/v1/test/depara/${selectedProduct.id}?table=${getCurrentTableName()}`, {
        sku: editForm.sku.trim(),
        company: editForm.company.trim()
      });
      
      if (response.data.success) {
        // Update the product in the list
        setProducts(products.map(p => 
          p.id === selectedProduct.id 
            ? { ...p, sku: editForm.sku.trim(), company: editForm.company.trim() }
            : p
        ));
        setSelectedProduct({ ...selectedProduct, sku: editForm.sku.trim(), company: editForm.company.trim() });
        setIsEditing(false);
        alert('Produto atualizado com sucesso!');
      } else {
        alert(response.data.message || 'Erro ao atualizar produto');
      }
    } catch (err: any) {
      alert(err.response?.data?.message || 'Erro ao atualizar produto');
    }
  };

  const handleDelete = async () => {
    if (!selectedProduct) return;
    
    // Confirmação mais específica para delete
    if (!confirm(`Tem certeza que deseja deletar o produto "${selectedProduct.id}"?\n\nEsta ação não pode ser desfeita!`)) {
      return;
    }
    
    try {
      const response = await axios.delete(`/api/v1/test/depara/${selectedProduct.id}?table=${getCurrentTableName()}`);
      
      if (response.data.success) {
        setProducts(products.filter(p => p.id !== selectedProduct.id));
        setIsModalOpen(false);
        setSelectedProduct(null);
        alert('Produto deletado com sucesso!');
      } else {
        alert(response.data.message || 'Erro ao deletar produto');
      }
    } catch (err: any) {
      alert(err.response?.data?.message || 'Erro ao deletar produto');
    }
  };

  // Função auxiliar para formatar valores de frete
  const formatShippingCost = (cost: any): string => {
    if (!cost || typeof cost !== 'object') return 'R$ 0,00';
    
    // Se for um objeto sql.NullFloat64
    if (cost.Valid && typeof cost.Float64 === 'number') {
      return `R$ ${cost.Float64.toFixed(2).replace('.', ',')}`;
    }
    
    // Se for um número simples
    if (typeof cost === 'number') {
      return `R$ ${cost.toFixed(2).replace('.', ',')}`;
    }
    
    return 'R$ 0,00';
  };

  const handleCreate = async () => {
    // Validação dos campos obrigatórios
    if (!createForm.id.trim() || !createForm.sku.trim() || !createForm.company.trim()) {
      alert('ID, SKU e Company são campos obrigatórios.');
      return;
    }
    
    // Confirmação antes de criar
    if (!confirm(`Tem certeza que deseja criar o produto "${createForm.id}"?`)) {
      return;
    }
    
    try {
      const response = await axios.post('/api/v1/test/depara', {
        table_name: getCurrentTableName(),
        id: createForm.id.trim(),
        sku: createForm.sku.trim(),
        company: createForm.company.trim(),
        mlbu: createForm.mlbu.trim() || createForm.id.trim(),
        type: createForm.type.trim() || 'product'
      });
      
      if (response.data.success) {
        setShowCreateForm(false);
        setCreateForm({ id: '', sku: '', company: '', mlbu: '', type: '' });
        alert('Produto criado com sucesso!');
        // Optionally refresh the search
        if (searchQuery) {
          handleSearch();
        }
      } else {
        alert(response.data.message || 'Erro ao criar produto');
      }
    } catch (err: any) {
      alert(err.response?.data?.message || 'Erro ao criar produto');
    }
  };

  return (
    <div className="p-6">
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-gray-900 mb-2">DePara</h1>
        <p className="text-gray-600">Gerencie produtos e integrações</p>
      </div>

      {/* Search Section */}
      <div className="bg-white rounded-lg shadow p-6 mb-6">
        <div className="space-y-4">
          {/* Dynamic Table Selection */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Configuração da Tabela
            </label>
            <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
              {/* Empresa */}
              <div>
                <label className="block text-xs font-medium text-gray-600 mb-1">
                  Empresa
                </label>
                <select
                  value={selectedEmpresa}
                  onChange={(e) => setSelectedEmpresa(e.target.value)}
                  className="w-full px-3 py-2 text-sm border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                >
                  {tableOptions.empresa.length > 0 ? (
                    tableOptions.empresa.map((empresa) => (
                      <option key={empresa} value={empresa}>
                        {empresa}
                      </option>
                    ))
                  ) : (
                    <option value="">Carregando...</option>
                  )}
                </select>
              </div>
              
              {/* Conta */}
              <div>
                <label className="block text-xs font-medium text-gray-600 mb-1">
                  Conta
                </label>
                <select
                  value={selectedConta}
                  onChange={(e) => setSelectedConta(e.target.value)}
                  className="w-full px-3 py-2 text-sm border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                >
                  {tableOptions.conta.length > 0 ? (
                    tableOptions.conta.map((conta) => (
                      <option key={conta} value={conta}>
                        {conta}
                      </option>
                    ))
                  ) : (
                    <option value="">Carregando...</option>
                  )}
                </select>
              </div>
              
              {/* Marketplace */}
              <div>
                <label className="block text-xs font-medium text-gray-600 mb-1">
                  Marketplace
                </label>
                <select
                  value={selectedMarketplace}
                  onChange={(e) => setSelectedMarketplace(e.target.value)}
                  className="w-full px-3 py-2 text-sm border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                >
                  {tableOptions.marketplace.length > 0 ? (
                    tableOptions.marketplace.map((marketplace) => (
                      <option key={marketplace} value={marketplace}>
                        {marketplace}
                      </option>
                    ))
                  ) : (
                    <option value="">Carregando...</option>
                  )}
                </select>
              </div>
            </div>
            
          </div>

          {/* Search and Actions Row */}
          <div className="flex flex-col sm:flex-row gap-3">
            {/* Search Input */}
            <div className="flex-1">
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Pesquisar (ID, MLBU ou SKU)
              </label>
              <div className="flex gap-2">
                <input
                  type="text"
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  placeholder="Digite ID, MLBU ou SKU..."
                  className="flex-1 px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  onKeyPress={(e) => e.key === 'Enter' && handleSearch()}
                />
                <button
                  onClick={() => handleSearch()}
                  disabled={loading || !searchQuery.trim()}
                  className="w-10 h-10 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50 flex items-center justify-center"
                  title="Pesquisar"
                >
                  <Search size={18} />
                </button>
              </div>
            </div>

            {/* Add Button */}
            <div className="flex items-end">
              <button
                onClick={() => setShowCreateForm(true)}
                className="w-10 h-10 bg-green-600 text-white rounded-md hover:bg-green-700 flex items-center justify-center"
                title="Adicionar produto"
              >
                <Plus size={18} />
              </button>
            </div>
          </div>
        </div>

        {error && (
          <div className="mt-4 p-3 bg-red-100 border border-red-400 text-red-700 rounded">
            {error}
          </div>
        )}
      </div>

      {/* Results Grid */}
        {products && products.length > 0 && (
            <div className="space-y-6">
              {/* Informações de resultados */}
              <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
                <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2">
                  <div className="text-sm text-blue-800">
                    <span className="font-semibold">{totalCount}</span> resultado{totalCount !== 1 ? 's' : ''} encontrado{totalCount !== 1 ? 's' : ''}
                  </div>
                  <div className="text-sm text-blue-600">
                    Página {currentPage} de {totalPages}
                  </div>
                </div>
              </div>
              
              {/* Grid de produtos - 3 linhas × 5 colunas */}
              <div className="grid grid-cols-5 gap-2.5 max-w-5xl mx-auto">
          {products.map((product) => (
            <div
              key={product.id}
              onClick={() => handleProductClick(product)}
              className="bg-white rounded-lg shadow hover:shadow-lg transition-shadow cursor-pointer overflow-hidden"
            >
              {/* Single Image */}
              <div className="aspect-square relative">
                {product.pictures && product.pictures.length > 0 ? (
                  <img
                    src={product.pictures[0]}
                    alt={product.sku}
                    className="w-full h-full object-cover"
                  />
                ) : (
                  <div className="w-full h-full bg-gray-200 flex items-center justify-center text-gray-500 text-xs">
                    Sem imagem
                  </div>
                )}
              </div>
              
              {/* Product Info */}
              <div className="p-2.5">
                <h3 className="font-bold text-sm text-gray-900 mb-1 truncate">
                  {product.id}
                </h3>
                <p className="text-xs text-gray-600 mb-1 truncate">{product.mlbu}</p>
                <div className="text-xs text-gray-500">
                  <div className="truncate">Empresa: {product.company}</div>
                  <div className="truncate">Type: {product.type}</div>
                </div>
              </div>
            </div>
          ))}
        </div>
        
        {/* Pagination Controls */}
        {totalPages > 1 && (
          <div className="flex items-center justify-center space-x-2 mt-6">
            <button
              onClick={() => handleSearch(currentPage - 1)}
              disabled={currentPage === 1 || loading}
              className="px-3 py-2 text-sm font-medium text-gray-500 bg-white border border-gray-300 rounded-md hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              Anterior
            </button>
            
            {/* Page Numbers */}
            <div className="flex space-x-1">
              {Array.from({ length: Math.min(5, totalPages) }, (_, i) => {
                let pageNum: number;
                if (totalPages <= 5) {
                  pageNum = i + 1;
                } else if (currentPage <= 3) {
                  pageNum = i + 1;
                } else if (currentPage >= totalPages - 2) {
                  pageNum = totalPages - 4 + i;
                } else {
                  pageNum = currentPage - 2 + i;
                }
                
                return (
                  <button
                    key={pageNum}
                    onClick={() => handleSearch(pageNum)}
                    disabled={loading}
                    className={`px-3 py-2 text-sm font-medium rounded-md ${
                      currentPage === pageNum
                        ? 'bg-blue-600 text-white'
                        : 'text-gray-500 bg-white border border-gray-300 hover:bg-gray-50'
                    } disabled:opacity-50 disabled:cursor-not-allowed`}
                  >
                    {pageNum}
                  </button>
                );
              })}
            </div>
            
            <button
              onClick={() => handleSearch(currentPage + 1)}
              disabled={currentPage === totalPages || loading}
              className="px-3 py-2 text-sm font-medium text-gray-500 bg-white border border-gray-300 rounded-md hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              Próximo
            </button>
          </div>
        )}
      </div>
      )}

      {/* Create Form Modal */}
      {showCreateForm && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg p-6 w-full max-w-md">
            <h3 className="text-lg font-semibold mb-4">Adicionar Novo Produto</h3>
            
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  ID * (obrigatório)
                </label>
                <input
                  type="text"
                  value={createForm.id}
                  onChange={(e) => setCreateForm({ ...createForm, id: e.target.value })}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="Ex: MLB123456789"
                />
              </div>
              
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  SKU * (obrigatório)
                </label>
                <input
                  type="text"
                  value={createForm.sku}
                  onChange={(e) => setCreateForm({ ...createForm, sku: e.target.value })}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="Ex: SKU001"
                />
              </div>
              
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Company * (obrigatório)
                </label>
                <input
                  type="text"
                  value={createForm.company}
                  onChange={(e) => setCreateForm({ ...createForm, company: e.target.value })}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="Ex: Amazonas"
                />
              </div>
              
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  MLBU (opcional)
                </label>
                <input
                  type="text"
                  value={createForm.mlbu}
                  onChange={(e) => setCreateForm({ ...createForm, mlbu: e.target.value })}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="Deixe vazio para usar o ID"
                />
              </div>
              
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Type (opcional)
                </label>
                <input
                  type="text"
                  value={createForm.type}
                  onChange={(e) => setCreateForm({ ...createForm, type: e.target.value })}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="Padrão: product"
                />
              </div>
            </div>
            
            <div className="flex gap-3 mt-6">
              <button
                onClick={handleCreate}
                className="flex-1 px-4 py-2 bg-green-600 text-white rounded-md hover:bg-green-700"
              >
                Criar
              </button>
              <button
                onClick={() => setShowCreateForm(false)}
                className="flex-1 px-4 py-2 bg-gray-300 text-gray-700 rounded-md hover:bg-gray-400"
              >
                Cancelar
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Product Detail Modal */}
      {isModalOpen && selectedProduct && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
          <div className="bg-white rounded-lg p-6 w-full max-w-4xl max-h-[90vh] overflow-y-auto">
            <div className="flex justify-between items-start mb-6">
              <h3 className="text-xl font-semibold">Detalhes do Anúncio</h3>
              <button
                onClick={() => setIsModalOpen(false)}
                className="text-gray-500 hover:text-gray-700"
              >
                <X size={24} />
              </button>
            </div>
            
            {/* Layout Vertical - Imagem em cima, informações embaixo */}
            <div className="space-y-6">
              {/* Image Slider - Quadrado centralizado com fundo branco */}
              <div className="flex justify-center">
                <div className="w-80 h-80 bg-white border border-gray-200 rounded-lg overflow-hidden shadow-sm">
                  <ImageSlider 
                    images={selectedProduct.pictures} 
                    className="h-full w-full"
                    showDots={true}
                    showArrows={true}
                  />
                </div>
              </div>
              
              {/* Product Details - Grid responsivo */}
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">SKU</label>
                  <input
                    type="text"
                    value={isEditing ? editForm.sku : selectedProduct.sku}
                    onChange={(e) => setEditForm({ ...editForm, sku: e.target.value })}
                    disabled={!isEditing}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md bg-gray-50"
                  />
                </div>
                
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">ID</label>
                  <input
                    type="text"
                    value={selectedProduct.id}
                    disabled
                    className="w-full px-3 py-2 border border-gray-300 rounded-md bg-gray-50"
                  />
                </div>
                
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">MLBU</label>
                  <input
                    type="text"
                    value={selectedProduct.mlbu}
                    disabled
                    className="w-full px-3 py-2 border border-gray-300 rounded-md bg-gray-50"
                  />
                </div>
                
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Empresa</label>
                  <input
                    type="text"
                    value={isEditing ? editForm.company : selectedProduct.company}
                    onChange={(e) => setEditForm({ ...editForm, company: e.target.value })}
                    disabled={!isEditing}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md bg-gray-50"
                  />
                </div>
                
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Type</label>
                  <input
                    type="text"
                    value={selectedProduct.type}
                    disabled
                    className="w-full px-3 py-2 border border-gray-300 rounded-md bg-gray-50"
                  />
                </div>
                
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Atualizado em</label>
                  <input
                    type="text"
                    value={new Date(selectedProduct.updated_at).toLocaleString('pt-BR')}
                    disabled
                    className="w-full px-3 py-2 border border-gray-300 rounded-md bg-gray-50"
                  />
                </div>
                
                {/* Permalink - Ocupa 2 colunas para não cortar */}
                <div className="md:col-span-2 lg:col-span-3">
                  <label className="block text-sm font-medium text-gray-700 mb-1">URL do Anúncio</label>
                  <input
                    type="text"
                    value={selectedProduct.permalink}
                    disabled
                    className="w-full px-3 py-2 border border-gray-300 rounded-md bg-gray-50 text-sm"
                  />
                </div>
                
                {/* Frete - Grid de 3 colunas */}
                <div className="md:col-span-2 lg:col-span-3">
                  <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">Frete Lento</label>
                      <input
                        type="text"
                        value={formatShippingCost(selectedProduct.ship_cost_slow)}
                        disabled
                        className="w-full px-3 py-2 border border-gray-300 rounded-md bg-gray-50 text-sm"
                      />
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">Frete Normal</label>
                      <input
                        type="text"
                        value={formatShippingCost(selectedProduct.ship_cost_standard)}
                        disabled
                        className="w-full px-3 py-2 border border-gray-300 rounded-md bg-gray-50 text-sm"
                      />
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">Frete Rápido</label>
                      <input
                        type="text"
                        value={formatShippingCost(selectedProduct.ship_cost_nextday)}
                        disabled
                        className="w-full px-3 py-2 border border-gray-300 rounded-md bg-gray-50 text-sm"
                      />
                    </div>
                  </div>
                </div>
              </div>
            </div>
            
            {/* Action Buttons */}
            <div className="flex gap-3 mt-6">
              {!isEditing ? (
                <>
                  <button
                    onClick={handleEdit}
                    className="flex-1 px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 flex items-center justify-center gap-2"
                  >
                    <Edit size={16} />
                    Editar
                  </button>
                  <button
                    onClick={handleDelete}
                    className="flex-1 px-4 py-2 bg-red-600 text-white rounded-md hover:bg-red-700 flex items-center justify-center gap-2"
                  >
                    <Trash2 size={16} />
                    Deletar
                  </button>
                </>
              ) : (
                <>
                  <button
                    onClick={handleUpdate}
                    className="flex-1 px-4 py-2 bg-green-600 text-white rounded-md hover:bg-green-700 flex items-center justify-center gap-2"
                  >
                    <Check size={16} />
                    Atualizar
                  </button>
                  <button
                    onClick={() => setIsEditing(false)}
                    className="flex-1 px-4 py-2 bg-gray-300 text-gray-700 rounded-md hover:bg-gray-400"
                  >
                    Cancelar
                  </button>
                </>
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  );
}


