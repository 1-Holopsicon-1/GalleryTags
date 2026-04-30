import {
  GetInboxPath as _GetInboxPath,
  SetInboxPath as _SetInboxPath,
  OpenDirectoryDialog as _OpenDirectoryDialog,
} from '../../wailsjs/go/main/App';

export async function getInboxPath(): Promise<string> {
  return _GetInboxPath();
}

export async function setInboxPath(path: string): Promise<void> {
  await _SetInboxPath(path);
}

export async function openDirectoryDialog(): Promise<string> {
  return _OpenDirectoryDialog();
}
