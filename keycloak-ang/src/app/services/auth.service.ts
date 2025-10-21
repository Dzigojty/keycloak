// auth.service.ts
import { Injectable } from '@angular/core';
import { KeycloakService } from 'keycloak-angular';
import { KeycloakAdminService } from 'src/app/services/keycloak.service';

@Injectable({
  providedIn: 'root'
})
export class AuthService {
  keycloak: any;
  constructor(
    private keycloakService: KeycloakService,
    private keycloakAdmin: KeycloakAdminService
  ) {}

  // Регистрация пользователя
  async register(userData: any): Promise<{success: boolean, message: string}> {
    try {
      // 1. Создаем пользователя в Keycloak
      await this.keycloakAdmin.createUser(userData);
      
      // 2. Автоматически логиним пользователя после регистрации
      await this.keycloakService.login({
        username: userData.username,
        password: userData.password,
        redirectUri: window.location.origin + '/dashboard'
      });

      return { success: true, message: 'Регистрация успешна!' };
    } catch (error: any) {
      console.error('Registration error:', error);
      
      if (error.status === 409) {
        return { success: false, message: 'Пользователь с таким email или логином уже существует' };
      } else {
        return { success: false, message: 'Ошибка регистрации. Попробуйте позже.' };
      }
    }
  }

  // Проверка авторизации
  async canActivate(): Promise<boolean> {
    return await this.keycloak.isLoggedIn();
  }

  // Выход
  logout(): void {
    this.keycloakService.logout();
  }
}