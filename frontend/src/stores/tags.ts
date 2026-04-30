import { writable } from 'svelte/store';
import type { Tag } from '../types';
import * as api from '../api/index';

export const tags = writable<Tag[]>([]);

export async function loadTags() {
  const result = await api.getTags();
  tags.set(result);
}
