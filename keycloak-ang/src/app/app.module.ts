// app.module.ts
import { NgModule, APP_INITIALIZER } from '@angular/core';
import { BrowserModule } from '@angular/platform-browser';
import { CommonModule } from '@angular/common';
import { ReactiveFormsModule, FormsModule } from '@angular/forms';
import { HttpClientModule } from '@angular/common/http';
import { RouterModule } from '@angular/router';
import { KeycloakAngularModule, KeycloakService } from 'keycloak-angular';

import { AppComponent } from './app.component';
import { RegistrationComponent } from './components/registration/registration.component';
import { EntranceComponent } from './components/entrance/entrance.component';
import { LoginComponent } from './components/login/login.component';
import { DashboardComponent } from './components/dashboard/dashboard.component';

import { routes } from './app.routes';

// Функция для инициализации Keycloak
function initializeKeycloak(keycloak: KeycloakService) {
  return () =>
    keycloak.init({
      config: {
        url: 'http://localhost:8080',
        realm: 'my-app',
        clientId: 'angular-app'
      },
      initOptions: {
        // ⚠️ ИЗМЕНИТЕ ЭТУ НАСТРОЙКУ ⚠️
        onLoad: 'check-sso', // Вместо 'login-required'
        checkLoginIframe: false,
        pkceMethod: 'S256'
      }
    }).then((authenticated) => {
      console.log('Keycloak инициализирован, пользователь аутентифицирован:', authenticated);
      if (authenticated) {
        console.log('Пользователь уже вошел - можно перенаправить на dashboard');
      } else {
        console.log('Пользователь не аутентифицирован - показываем логин форму');
      }
    }).catch(error => {
      console.error('Ошибка инициализации Keycloak:', error);
    });
}

@NgModule({
  declarations: [
    AppComponent,
    RegistrationComponent,
    EntranceComponent,
    LoginComponent,
    DashboardComponent
  ],
  imports: [
    BrowserModule,
    ReactiveFormsModule,
    CommonModule,
    FormsModule,
    HttpClientModule,
    RouterModule.forRoot(routes),
    KeycloakAngularModule
  ],
  providers: [
    {
      provide: APP_INITIALIZER,
      useFactory: initializeKeycloak,
      multi: true,
      deps: [KeycloakService]
    }
  ],
  bootstrap: [AppComponent]
})
export class AppModule { }