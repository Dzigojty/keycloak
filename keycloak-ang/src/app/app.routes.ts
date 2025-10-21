// app.routes.ts
import { Routes } from '@angular/router';
import { HomeComponent } from './components/home/home.component';
import { EntranceComponent } from './components/entrance/entrance.component';
import { LoginComponent } from './components/login/login.component';
import { RegistrationComponent } from './components/registration/registration.component';
import { SucsessComponent } from './components/sucsess/sucsess.component';
import { DashboardComponent } from 'src/app/components/dashboard/dashboard.component';
import { authGuard } from 'src/app/auth.guard';

export const routes: Routes = [
  { path: '', component: HomeComponent },
  { path: 'entrance', component: EntranceComponent },
  { path: 'login', component: LoginComponent },
  { path: 'registration', component: RegistrationComponent },
  { path: 'register', component: RegistrationComponent },
  { path: 'sucsess', component: SucsessComponent },
    { 
    path: 'dashboard', 
    component: DashboardComponent,
    canActivate: [authGuard] 
  },
  { path: '', redirectTo: '/login', pathMatch: 'full' },
  { path: '**', redirectTo: '/login' },
  { path: '**', redirectTo: '' }
];
