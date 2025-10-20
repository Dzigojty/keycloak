// src/app/components/login/login.component.ts
import { Component, OnInit } from '@angular/core';
import { Router } from '@angular/router';
import { KeycloakService } from 'src/app/services/keycloak.service';

@Component({
  selector: 'app-login',
  templateUrl: './login.component.html',
  styleUrls: ['./login.component.css']
})
export class LoginComponent implements OnInit {
  title = 'Вход в систему';
  
  // Свойства формы
  name = '';
  password = '';
  isLoading = false;
  errorMessage = '';
  
  constructor(
    private keycloakService: KeycloakService,
    private router: Router
  ) {}

  ngOnInit(): void {
    // Проверяем, авторизован ли пользователь при загрузке компонента
    if (this.keycloakService.isLoggedIn()) {
      this.redirectAfterLogin();
    }
  }

  // Метод для входа через Keycloak (стандартный flow с редиректом)
  onKeycloakLogin(): void {
    this.isLoading = true;
    this.errorMessage = '';
    
    // Добавим таймаут на случай, если Keycloak недоступен
    setTimeout(() => {
      if (this.isLoading) {
        this.isLoading = false;
        this.errorMessage = 'Keycloak сервер недоступен. Проверьте, запущен ли контейнер Keycloak.';
      }
    }, 5000);
    
    this.keycloakService.login();
  }

  // Метод для прямого входа (если нужно использовать свою форму)
  onDirectLogin(): void {
    if (!this.name || !this.password) {
      this.errorMessage = 'Пожалуйста, заполните все поля';
      return;
    }

    this.isLoading = true;
    this.errorMessage = '';

    this.keycloakService.loginDirect(this.name, this.password).subscribe({
      next: (response) => {
        this.isLoading = false;
        console.log('Прямой вход успешен:', response);
        this.redirectAfterLogin();
      },
      error: (error) => {
        this.isLoading = false;
        console.error('Ошибка прямого входа:', error);
        this.handleLoginError(error);
      }
    });
  }

  // Метод для перенаправления после успешного входа
  private redirectAfterLogin() {
    this.router.navigate(['/admin']);
  }

  // Метод для обработки ошибок авторизации
  private handleLoginError(error: any) {
    if (error.error && error.error.error === 'invalid_grant') {
      this.errorMessage = 'Неверное имя пользователя или пароль';
    } else if (error.status === 401) {
      this.errorMessage = 'Ошибка аутентификации';
    } else if (error.status === 403) {
      this.errorMessage = 'Доступ запрещен';
    } else if (error.status === 0) {
      this.errorMessage = 'Не удается подключиться к серверу Keycloak';
    } else if (error.status >= 500) {
      this.errorMessage = 'Ошибка сервера Keycloak. Попробуйте позже.';
    } else {
      this.errorMessage = 'Произошла ошибка при авторизации';
    }
  }

  // Дополнительные методы
  onForgotPassword() {
    this.keycloakService.forgotPassword();
  }

  onRegister() {
    this.keycloakService.register();
  }

  // Очистка ошибки при изменении полей
  onInputChange() {
    if (this.errorMessage) {
      this.errorMessage = '';
    }
  }
}