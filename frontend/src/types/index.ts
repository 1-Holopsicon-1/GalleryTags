export interface Tag {
  id: number;
  name: string;
  type: 'label' | 'folder';
  folder: string;
  color: string;
  hotkey: string;
}

export interface FileInfo {
  path: string;
  name: string;
  type: 'image' | 'video';
  modTime: string;
  birthTime: string;
}

export interface ApplyResult {
  newPath: string;
  moved: boolean;
}
