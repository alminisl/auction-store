import { Link } from 'react-router-dom';
import { useTranslation } from 'react-i18next';

export default function Footer() {
  const { t } = useTranslation();

  return (
    <footer className="border-t bg-muted/30">
      <div className="container-custom py-12">
        <div className="grid grid-cols-1 md:grid-cols-4 gap-8">
          {/* Brand */}
          <div className="col-span-1 md:col-span-2">
            <Link to="/" className="inline-block mb-4">
              <img src="/logo.svg" alt="Pirates&Magic" className="h-10" />
            </Link>
            <p className="text-sm text-muted-foreground max-w-sm">
              {t('footer.tagline')}
            </p>
          </div>

          {/* Quick Links */}
          <div>
            <h3 className="font-semibold mb-4">{t('footer.quickLinks')}</h3>
            <ul className="space-y-2 text-sm text-muted-foreground">
              <li>
                <Link to="/auctions" className="hover:text-foreground transition-colors">
                  {t('footer.browseAuctions')}
                </Link>
              </li>
              <li>
                <Link to="/categories" className="hover:text-foreground transition-colors">
                  {t('footer.categories')}
                </Link>
              </li>
              <li>
                <Link to="/auctions/create" className="hover:text-foreground transition-colors">
                  {t('footer.sellItem')}
                </Link>
              </li>
            </ul>
          </div>

          {/* Support */}
          <div>
            <h3 className="font-semibold mb-4">{t('footer.support')}</h3>
            <ul className="space-y-2 text-sm text-muted-foreground">
              <li>
                <Link to="/help" className="hover:text-foreground transition-colors">
                  {t('footer.helpCenter')}
                </Link>
              </li>
              <li>
                <Link to="/contact" className="hover:text-foreground transition-colors">
                  {t('footer.contactUs')}
                </Link>
              </li>
              <li>
                <Link to="/terms" className="hover:text-foreground transition-colors">
                  {t('footer.termsOfService')}
                </Link>
              </li>
              <li>
                <Link to="/privacy" className="hover:text-foreground transition-colors">
                  {t('footer.privacyPolicy')}
                </Link>
              </li>
            </ul>
          </div>
        </div>

        <div className="mt-8 pt-8 border-t text-center text-sm text-muted-foreground">
          <p>&copy; {new Date().getFullYear()} {t('common.appName')}. {t('footer.allRightsReserved')}</p>
        </div>
      </div>
    </footer>
  );
}
