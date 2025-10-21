import { Component } from '@angular/core';

@Component({
  selector: 'app-home',
  templateUrl: './home.component.html',
  styleUrls: ['./home.component.css']
})
export class HomeComponent {
  title = 'Добро пожаловать на главную страницу!';
  fileContent: string = '';
  isContentVisible: boolean = false;

  onFileSelected(event: any): void {
    const file: File = event.target.files[0];
    
    if (file && file.type === 'text/plain') {
      this.readFile(file);
    } else if (file) {
      alert('Пожалуйста, выберите файл в формате TXT');
    }
  }

  private readFile(file: File): void {
    const reader = new FileReader();
    
    reader.onload = (e: ProgressEvent<FileReader>) => {
      this.fileContent = (e.target as FileReader).result as string;
      this.isContentVisible = true;
    };
    
    reader.onerror = () => {
      alert('Ошибка при чтении файла');
    };
    
    reader.readAsText(file, 'UTF-8');
  }

  triggerFileInput(): void {
    // В Angular лучше использовать ViewChild вместо прямого доступа к DOM
    const fileInput = document.getElementById('fileInput') as HTMLInputElement;
    if (fileInput) {
      fileInput.click();
    }
  }
}