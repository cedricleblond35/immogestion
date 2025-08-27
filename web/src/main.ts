
import { bootstrapApplication } from '@angular/platform-browser';
import { provideHttpClient, withInterceptorsFromDi, withFetch } from '@angular/common/http';
import { provideRouter } from '@angular/router';
import { importProvidersFrom } from '@angular/core';

import { App } from './app/app';
import { routes } from './app/app.routes'; // Vos routes

bootstrapApplication(App, {
  providers: [
    // Configuration du HttpClient
    provideHttpClient(
      withInterceptorsFromDi(), // Pour utiliser les interceptors classiques si nécessaire
      withFetch() // Utilise l'API Fetch moderne (optionnel, recommandé)
    ),
    
    // Configuration du routeur
    provideRouter(routes),
    
    // Autres providers si nécessaire
    // provideAnimations(), // Pour les animations
    
    // Providers personnalisés
    // { provide: API_BASE_URL, useValue: 'http://localhost:8080/api' }
  ]
}).catch(err => console.error(err));