import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { useEffect } from 'react';
import { useAuthStore } from './store';
import { Layout } from './components/layout';

// Pages
import Home from './pages/Home';
import Login from './pages/Login';
import Register from './pages/Register';
import Auctions from './pages/Auctions';
import AuctionDetail from './pages/AuctionDetail';
import CreateAuction from './pages/CreateAuction';
import EditAuction from './pages/EditAuction';
import MyAuctions from './pages/MyAuctions';
import MyPurchases from './pages/MyPurchases';
import Watchlist from './pages/Watchlist';
import Categories from './pages/Categories';

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 1000 * 60 * 5, // 5 minutes
      retry: 1,
    },
  },
});

// Protected Route wrapper
function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const { isAuthenticated } = useAuthStore();

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }

  return <>{children}</>;
}

export default function App() {
  const { fetchUser, isAuthenticated } = useAuthStore();

  useEffect(() => {
    if (isAuthenticated) {
      fetchUser();
    }
  }, [isAuthenticated, fetchUser]);

  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <Routes>
          <Route path="/" element={<Layout />}>
            {/* Public routes */}
            <Route index element={<Home />} />
            <Route path="login" element={<Login />} />
            <Route path="register" element={<Register />} />

            {/* Auction routes */}
            <Route path="auctions" element={<Auctions />} />
            <Route path="auctions/:id" element={<AuctionDetail />} />
            <Route path="categories" element={<Categories />} />

            {/* Protected routes */}
            <Route
              path="auctions/create"
              element={
                <ProtectedRoute>
                  <CreateAuction />
                </ProtectedRoute>
              }
            />
            <Route
              path="auctions/:id/edit"
              element={
                <ProtectedRoute>
                  <EditAuction />
                </ProtectedRoute>
              }
            />
            <Route
              path="profile"
              element={
                <ProtectedRoute>
                  <div className="container-custom py-8">Profile page coming soon</div>
                </ProtectedRoute>
              }
            />
            <Route
              path="my-auctions"
              element={
                <ProtectedRoute>
                  <MyAuctions />
                </ProtectedRoute>
              }
            />
            <Route
              path="my-bids"
              element={
                <ProtectedRoute>
                  <MyPurchases />
                </ProtectedRoute>
              }
            />
            <Route
              path="my-purchases"
              element={
                <ProtectedRoute>
                  <MyPurchases />
                </ProtectedRoute>
              }
            />
            <Route
              path="watchlist"
              element={
                <ProtectedRoute>
                  <Watchlist />
                </ProtectedRoute>
              }
            />
            <Route
              path="notifications"
              element={
                <ProtectedRoute>
                  <div className="container-custom py-8">Notifications page coming soon</div>
                </ProtectedRoute>
              }
            />

            {/* 404 */}
            <Route path="*" element={<div className="container-custom py-8">Page not found</div>} />
          </Route>
        </Routes>
      </BrowserRouter>
    </QueryClientProvider>
  );
}
