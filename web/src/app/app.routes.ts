import { Routes } from '@angular/router';
import { Home } from './features/public/home/home';
import { PublicLayout } from './features/layouts/public-layout/public-layout';
import { PrivateLayout } from './features/layouts/private-layout/private-layout';

export const routes: Routes = [
  // Layout Public
  {
    path: '',
    component: PublicLayout,
    //canActivate: [publicGuard], // si connecté → redirigé vers dashboard
    children: [
      { path: '', component: Home },
      {
        path: 'auth',
        loadChildren: () => import('./features/public/auth/auth.routes')
                           .then(m => m.AUTH_ROUTES),
      },
    ]
  },

  // Layout Privé
  {
    path: '',
    component: PrivateLayout,
    //canActivate: [authGuard], // si pas connecté → redirigé vers /auth/login
    children: [
      {
        path: 'dashboard',
        loadChildren: () => import('./features/private/dashboard/dashboard.routes')
                           .then(m => m.DASHBOARD_ROUTES),
      },
      {
        path: 'properties',
        loadChildren: () => import('./features/private/properties/properties.routes')
                           .then(m => m.PROPERTIES_ROUTES),
      },
    ]
  },
  
   
  // Catch-all
  { path: '**', redirectTo: '' }
];