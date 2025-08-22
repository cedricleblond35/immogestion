import { Routes } from '@angular/router';

export const routes: Routes = [
  {
    path: 'auth',
    loadChildren: () => import('./features/auth/auth.routes')
                        .then(m => m.AUTH_ROUTES),
  },
  {
    path: 'properties',
    loadChildren: () => import('./features/properties/properties.routes')
                        .then(m => m.PROPERTIES_ROUTES),
  },
  { path: '', redirectTo: 'dashboard', pathMatch: 'full' },
  { path: '**', redirectTo: 'dashboard' }, // (optionnel pour les 404)
];