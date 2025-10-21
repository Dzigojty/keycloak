// app.module.ts
import { NgModule, APP_INITIALIZER } from '@angular/core';
import { BrowserModule } from '@angular/platform-browser';
import { ReactiveFormsModule, FormsModule } from '@angular/forms'; // ← ДОБАВЬТЕ FormsModule
import { HttpClientModule } from '@angular/common/http';
import { RouterModule } from '@angular/router';
import { KeycloakAngularModule, KeycloakService } from 'keycloak-angular';

import { AppComponent } from './app.component';
import { RegistrationComponent } from './components/registration/registration.component';
import { LoginComponent } from './components/login/login.component'; // ← ДОБАВЬТЕ ЭТОТ ИМПОРТ

import { routes } from './app.routes';

// Функция для инициализации Keycloak
function initializeKeycloak(keycloak: KeycloakService) {
  return () =>
    keycloak.init({
      config: {
        url: 'http://localhost:8080',
        realm: 'my-app', // ← Используйте ваш realm
        clientId: 'angular-app' // ← Создайте отдельного клиента для Angular
      },
      initOptions: {
        onLoad: 'login-required', // ← Измените на login-required для теста
        checkLoginIframe: false,
        pkceMethod: 'S256'
      }
    });
}

@NgModule({
  declarations: [
    AppComponent,
    RegistrationComponent,
    LoginComponent // ← ДОБАВЬТЕ ЭТО В DECLARATIONS
  ],
  imports: [
    BrowserModule,
    ReactiveFormsModule,
    FormsModule, // ← ДОБАВЬТЕ ЭТО ДЛЯ ngModel
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
