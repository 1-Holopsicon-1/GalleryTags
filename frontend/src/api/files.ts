import type { FileInfo, ApplyResult } from '../types';

import {
  ScanInbox as _ScanInbox,
  ApplyTags as _ApplyTags,
  TrashFile as _TrashFile,
} from '../../wailsjs/go/main/App';

function mapFileInfo(raw: any): FileInfo {
  return {
    path: raw.Path ?? raw.path,
    name: raw.Name ?? raw.name,
    type: raw.Type ?? raw.type,
    modTime: raw.ModTime ?? raw.modTime ?? '',
    birthTime: raw.BirthTime ?? raw.birthTime ?? '',
  };
}

export async function scanInbox(recursive: boolean, sortBy: string): Promise<FileInfo[]> {
  const raw = await _ScanInbox(recursive, sortBy);
  return Array.isArray(raw) ? raw.map(mapFileInfo) : [];
}

export async function applyTags(filePath: string, tagIDs: number[]): Promise<ApplyResult> {
  const raw: any = await _ApplyTags(filePath, tagIDs);
  return {
    newPath: raw.NewPath ?? raw.newPath ?? '',
    moved: raw.Moved ?? raw.moved ?? false,
  };
}

export async function trashFile(filePath: string): Promise<void> {
  await _TrashFile(filePath);
}
