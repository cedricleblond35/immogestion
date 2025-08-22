import { Routes } from '@angular/router';
import { PropertiesList } from './components/properties-list/properties-list';
import { PropertyCreate } from './components/property-create/property-create';
import { PropertyDetail } from './components/property-detail/property-detail';
import { PropertyDocuments } from './components/property-documents/property-documents';
import { PropertyTenants } from './components/property-tenants/property-tenants';
import { PropertyExpenses } from './components/property-expenses/property-expenses';

export const PROPERTIES_ROUTES: Routes = [
  { path: '', component: PropertiesList },          // /properties
  { path: 'create', component: PropertyCreate },   // /properties/create
  {
    path: ':id',
    component: PropertyDetail,                     // /properties/123
    children: [
      { path: 'documents', component: PropertyDocuments }, // /properties/123/documents
      { path: 'tenants', component: PropertyTenants },     // /properties/123/tenants
      { path: 'expenses', component: PropertyExpenses },   // /properties/123/expenses
    ],
  },
];
