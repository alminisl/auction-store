import { Link } from 'react-router-dom';
import { ArrowRight, Shield, Clock, Zap } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { Button } from '../components/common';

export default function Home() {
  const { t } = useTranslation();

  return (
    <div>
      {/* Hero Section */}
      <section className="bg-muted/50 py-20 lg:py-28">
        <div className="container-custom">
          <div className="max-w-3xl mx-auto text-center">
            <h1 className="text-4xl md:text-5xl lg:text-6xl font-bold text-foreground mb-6 leading-tight">
              {t('home.heroTitle')}{' '}
              <span className="text-primary">{t('home.heroTitleHighlight')}</span>
            </h1>
            <p className="text-lg text-muted-foreground mb-10 max-w-2xl mx-auto">
              {t('home.heroSubtitle')}
            </p>
            <div className="flex flex-col sm:flex-row gap-4 justify-center">
              <Link to="/auctions">
                <Button size="lg" className="w-full sm:w-auto px-8">
                  {t('home.browseAuctions')}
                  <ArrowRight className="ml-2 h-4 w-4" />
                </Button>
              </Link>
              <Link to="/auctions/create">
                <Button variant="outline" size="lg" className="w-full sm:w-auto px-8">
                  {t('home.startSelling')}
                </Button>
              </Link>
            </div>
          </div>
        </div>
      </section>

      {/* Features Section */}
      <section className="py-20">
        <div className="container-custom">
          <h2 className="text-3xl font-bold text-center mb-4">{t('home.whyChooseUs')}</h2>
          <p className="text-muted-foreground text-center mb-12 max-w-xl mx-auto">
            The trusted marketplace for collectors and enthusiasts
          </p>

          <div className="grid md:grid-cols-3 gap-8">
            <div className="text-center p-8 rounded-xl bg-background border border-border hover:shadow-lg transition-shadow">
              <div className="inline-flex items-center justify-center w-14 h-14 rounded-full bg-primary/10 mb-6">
                <Shield className="h-7 w-7 text-primary" />
              </div>
              <h3 className="text-xl font-semibold mb-3">{t('home.secureTransactions')}</h3>
              <p className="text-muted-foreground">
                {t('home.secureTransactionsDesc')}
              </p>
            </div>

            <div className="text-center p-8 rounded-xl bg-background border border-border hover:shadow-lg transition-shadow">
              <div className="inline-flex items-center justify-center w-14 h-14 rounded-full bg-primary/10 mb-6">
                <Clock className="h-7 w-7 text-primary" />
              </div>
              <h3 className="text-xl font-semibold mb-3">{t('home.realTimeBidding')}</h3>
              <p className="text-muted-foreground">
                {t('home.realTimeBiddingDesc')}
              </p>
            </div>

            <div className="text-center p-8 rounded-xl bg-background border border-border hover:shadow-lg transition-shadow">
              <div className="inline-flex items-center justify-center w-14 h-14 rounded-full bg-primary/10 mb-6">
                <Zap className="h-7 w-7 text-primary" />
              </div>
              <h3 className="text-xl font-semibold mb-3">{t('home.easyToUse')}</h3>
              <p className="text-muted-foreground">
                {t('home.easyToUseDesc')}
              </p>
            </div>
          </div>
        </div>
      </section>

      {/* CTA Section */}
      <section className="bg-primary py-16">
        <div className="container-custom text-center">
          <h2 className="text-3xl font-bold text-primary-foreground mb-4">
            {t('home.readyToStart')}
          </h2>
          <p className="text-primary-foreground/80 mb-8 max-w-xl mx-auto">
            {t('home.joinThousands')}
          </p>
          <Link to="/register">
            <Button variant="secondary" size="lg" className="px-8">
              {t('home.createFreeAccount')}
            </Button>
          </Link>
        </div>
      </section>
    </div>
  );
}
