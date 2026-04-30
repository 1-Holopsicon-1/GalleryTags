import type { Tag } from '../types';

import {
  GetTags as _GetTags,
  CreateTag as _CreateTag,
  UpdateTag as _UpdateTag,
  DeleteTag as _DeleteTag,
  GetFileTags as _GetFileTags,
  GetFilteredFiles as _GetFilteredFiles,
} from '../../wailsjs/go/main/App';

function mapTag(raw: any): Tag {
  return {
    id: raw.ID ?? raw.id,
    name: raw.Name ?? raw.name,
    type: raw.Type ?? raw.type,
    folder: raw.Folder ?? raw.folder ?? '',
    color: raw.Color ?? raw.color ?? '#4a9eff',
    hotkey: raw.Hotkey ?? raw.hotkey ?? '',
  };
}

export async function getTags(): Promise<Tag[]> {
  const raw = await _GetTags();
  return Array.isArray(raw) ? raw.map(mapTag) : [];
}

export async function createTag(name: string, tagType: string, folder: string, color: string, hotkey: string): Promise<Tag> {
  const raw = await _CreateTag(name, tagType, folder, color, hotkey);
  return mapTag(raw);
}

export async function updateTag(id: number, name: string, tagType: string, folder: string, color: string, hotkey: string): Promise<void> {
  await _UpdateTag(id, name, tagType, folder, color, hotkey);
}

export async function deleteTag(id: number): Promise<void> {
  await _DeleteTag(id);
}

export async function getFileTags(filePath: string): Promise<Tag[]> {
  const raw = await _GetFileTags(filePath);
  return Array.isArray(raw) ? raw.map(mapTag) : [];
}

export async function getFilteredFiles(tagIDs: number[]): Promise<string[]> {
  const raw = await _GetFilteredFiles(tagIDs);
  return Array.isArray(raw) ? raw : [];
}
