import { Link, useNavigate } from 'react-router-dom';
import { Search, Bell, User, Menu, LogOut, Settings, Package, Heart, ShoppingBag } from 'lucide-react';
import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useAuthStore } from '../../store';
import { Button, LanguageSwitcher } from '../common';
import { cn } from '../../utils';

export default function Header() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { user, isAuthenticated, logout } = useAuthStore();
  const [isMenuOpen, setIsMenuOpen] = useState(false);
  const [isUserMenuOpen, setIsUserMenuOpen] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();
    if (searchQuery.trim()) {
      navigate(`/auctions?search=${encodeURIComponent(searchQuery.trim())}`);
    }
  };

  const handleLogout = async () => {
    await logout();
    navigate('/');
  };

  return (
    <header className="sticky top-0 z-50 w-full bg-background border-b border-border">
      <div className="container-custom">
        <div className="flex h-16 items-center justify-between">
          {/* Logo */}
          <Link to="/" className="flex items-center">
            <img src="/logo.svg" alt="Pirates&Magic" className="h-10" />
          </Link>

          {/* Search Bar - Desktop */}
          <form onSubmit={handleSearch} className="hidden md:flex flex-1 max-w-lg mx-8">
            <div className="relative w-full">
              <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
              <input
                type="text"
                placeholder={t('common.searchPlaceholder')}
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="w-full h-10 rounded-lg border border-input bg-background pl-10 pr-4 text-sm
                  focus:outline-none focus:ring-2 focus:ring-primary focus:border-primary
                  transition-colors placeholder:text-muted-foreground"
              />
            </div>
          </form>

          {/* Navigation - Desktop */}
          <nav className="hidden md:flex items-center space-x-1">
            <Link
              to="/auctions"
              className="px-4 py-2 text-sm font-medium text-foreground hover:text-primary transition-colors"
            >
              {t('nav.browse')}
            </Link>
            <Link
              to="/categories"
              className="px-4 py-2 text-sm font-medium text-foreground hover:text-primary transition-colors"
            >
              {t('nav.categories')}
            </Link>

            {isAuthenticated ? (
              <>
                <Link
                  to="/auctions/create"
                  className="px-4 py-2 text-sm font-medium text-primary hover:text-primary/80 transition-colors"
                >
                  {t('nav.sell')}
                </Link>

                <div className="flex items-center space-x-1 ml-4 pl-4 border-l border-border">
                  <Link
                    to="/watchlist"
                    className="p-2 text-muted-foreground hover:text-foreground rounded-lg hover:bg-muted transition-colors"
                  >
                    <Heart className="h-5 w-5" />
                  </Link>

                  <Link
                    to="/notifications"
                    className="p-2 text-muted-foreground hover:text-foreground rounded-lg hover:bg-muted transition-colors"
                  >
                    <Bell className="h-5 w-5" />
                  </Link>

                  {/* User Menu */}
                  <div className="relative">
                    <button
                      onClick={() => setIsUserMenuOpen(!isUserMenuOpen)}
                      className="flex items-center space-x-2 p-1.5 hover:bg-muted rounded-lg transition-colors"
                    >
                      <div className="h-8 w-8 rounded-full bg-muted border border-border flex items-center justify-center overflow-hidden">
                        {user?.avatar_url ? (
                          <img src={user.avatar_url} alt="" className="h-8 w-8 rounded-full object-cover" />
                        ) : (
                          <User className="h-4 w-4 text-muted-foreground" />
                        )}
                      </div>
                    </button>

                    {isUserMenuOpen && (
                      <div className="absolute right-0 mt-2 w-56 rounded-lg border border-border bg-background shadow-lg">
                        <div className="p-2">
                          <div className="px-3 py-2 border-b border-border mb-1">
                            <p className="font-medium text-foreground">{user?.username}</p>
                            <p className="text-sm text-muted-foreground">{user?.email}</p>
                          </div>
                          <Link
                            to="/profile"
                            onClick={() => setIsUserMenuOpen(false)}
                            className="flex items-center px-3 py-2 text-sm text-foreground hover:bg-muted rounded-md transition-colors"
                          >
                            <User className="mr-3 h-4 w-4 text-muted-foreground" /> {t('nav.profile')}
                          </Link>
                          <Link
                            to="/my-auctions"
                            onClick={() => setIsUserMenuOpen(false)}
                            className="flex items-center px-3 py-2 text-sm text-foreground hover:bg-muted rounded-md transition-colors"
                          >
                            <Package className="mr-3 h-4 w-4 text-muted-foreground" /> {t('nav.myAuctions')}
                          </Link>
                          <Link
                            to="/my-purchases"
                            onClick={() => setIsUserMenuOpen(false)}
                            className="flex items-center px-3 py-2 text-sm text-foreground hover:bg-muted rounded-md transition-colors"
                          >
                            <ShoppingBag className="mr-3 h-4 w-4 text-muted-foreground" /> {t('nav.myPurchases')}
                          </Link>
                          <Link
                            to="/settings"
                            onClick={() => setIsUserMenuOpen(false)}
                            className="flex items-center px-3 py-2 text-sm text-foreground hover:bg-muted rounded-md transition-colors"
                          >
                            <Settings className="mr-3 h-4 w-4 text-muted-foreground" /> {t('nav.settings')}
                          </Link>
                          {user?.role === 'admin' && (
                            <Link
                              to="/admin"
                              onClick={() => setIsUserMenuOpen(false)}
                              className="flex items-center px-3 py-2 text-sm text-foreground hover:bg-muted rounded-md transition-colors"
                            >
                              <Settings className="mr-3 h-4 w-4 text-muted-foreground" /> {t('nav.admin')}
                            </Link>
                          )}
                          <div className="border-t border-border mt-1 pt-1">
                            <button
                              onClick={handleLogout}
                              className="flex w-full items-center px-3 py-2 text-sm text-destructive hover:bg-muted rounded-md transition-colors"
                            >
                              <LogOut className="mr-3 h-4 w-4" /> {t('nav.logout')}
                            </button>
                          </div>
                        </div>
                      </div>
                    )}
                  </div>
                </div>
              </>
            ) : (
              <div className="flex items-center space-x-3 ml-4 pl-4 border-l border-border">
                <Button variant="ghost" onClick={() => navigate('/login')}>
                  {t('nav.signIn')}
                </Button>
                <Button onClick={() => navigate('/register')}>
                  {t('nav.signUp')}
                </Button>
              </div>
            )}
            <div className="ml-2 pl-2 border-l border-border">
              <LanguageSwitcher />
            </div>
          </nav>

          {/* Mobile Menu Button */}
          <button
            onClick={() => setIsMenuOpen(!isMenuOpen)}
            className="md:hidden p-2 hover:bg-muted rounded-lg transition-colors"
          >
            <Menu className="h-6 w-6" />
          </button>
        </div>

        {/* Mobile Menu */}
        {isMenuOpen && (
          <div className="md:hidden py-4 border-t border-border">
            <form onSubmit={handleSearch} className="mb-4">
              <div className="relative">
                <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
                <input
                  type="text"
                  placeholder={t('common.searchPlaceholder')}
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className="w-full h-10 rounded-lg border border-input bg-background pl-10 pr-4 text-sm focus:outline-none focus:ring-2 focus:ring-primary"
                />
              </div>
            </form>
            <nav className="flex flex-col space-y-1">
              <Link
                to="/auctions"
                onClick={() => setIsMenuOpen(false)}
                className="px-3 py-2 text-sm font-medium text-foreground hover:bg-muted rounded-lg transition-colors"
              >
                {t('nav.browse')}
              </Link>
              <Link
                to="/categories"
                onClick={() => setIsMenuOpen(false)}
                className="px-3 py-2 text-sm font-medium text-foreground hover:bg-muted rounded-lg transition-colors"
              >
                {t('nav.categories')}
              </Link>
              {isAuthenticated ? (
                <>
                  <Link
                    to="/auctions/create"
                    onClick={() => setIsMenuOpen(false)}
                    className="px-3 py-2 text-sm font-medium text-primary hover:bg-muted rounded-lg transition-colors"
                  >
                    {t('nav.sell')}
                  </Link>
                  <Link
                    to="/profile"
                    onClick={() => setIsMenuOpen(false)}
                    className="px-3 py-2 text-sm font-medium text-foreground hover:bg-muted rounded-lg transition-colors"
                  >
                    {t('nav.profile')}
                  </Link>
                  <button
                    onClick={() => {
                      handleLogout();
                      setIsMenuOpen(false);
                    }}
                    className="px-3 py-2 text-sm font-medium text-destructive hover:bg-muted rounded-lg text-left transition-colors"
                  >
                    {t('nav.logout')}
                  </button>
                </>
              ) : (
                <>
                  <Link
                    to="/login"
                    onClick={() => setIsMenuOpen(false)}
                    className="px-3 py-2 text-sm font-medium text-foreground hover:bg-muted rounded-lg transition-colors"
                  >
                    {t('nav.signIn')}
                  </Link>
                  <Link
                    to="/register"
                    onClick={() => setIsMenuOpen(false)}
                    className="px-3 py-2 text-sm font-medium text-primary hover:bg-muted rounded-lg transition-colors"
                  >
                    {t('nav.signUp')}
                  </Link>
                </>
              )}
              <div className="px-3 py-2">
                <LanguageSwitcher />
              </div>
            </nav>
          </div>
        )}
      </div>
    </header>
  );
}
