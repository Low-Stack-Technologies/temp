import AsyncStorage from '@react-native-async-storage/async-storage';

const KEYS = {
  SERVER_URL: 'SERVER_URL',
  UPLOAD_HISTORY: 'UPLOAD_HISTORY',
};

export interface HistoryItem {
  id: string;
  filename: string;
  size: number;
  downloadUrl: string;
  uploadedAt: number;
}

export const saveServerUrl = async (url: string) => {
  try {
    await AsyncStorage.setItem(KEYS.SERVER_URL, url);
  } catch (e) {
    console.error('Failed to save server URL', e);
  }
};

export const getServerUrl = async (): Promise<string> => {
  try {
    const url = await AsyncStorage.getItem(KEYS.SERVER_URL);
    return url || 'https://temp.low-stack.tech';
  } catch (e) {
    console.error('Failed to get server URL', e);
    return 'https://temp.low-stack.tech';
  }
};

export const saveHistoryItem = async (item: HistoryItem) => {
  try {
    const history = await getHistory();
    const newHistory = [item, ...history];
    await AsyncStorage.setItem(KEYS.UPLOAD_HISTORY, JSON.stringify(newHistory));
  } catch (e) {
    console.error('Failed to save history item', e);
  }
};

export const getHistory = async (): Promise<HistoryItem[]> => {
  try {
    const json = await AsyncStorage.getItem(KEYS.UPLOAD_HISTORY);
    return json ? JSON.parse(json) : [];
  } catch (e) {
    console.error('Failed to get history', e);
    return [];
  }
};

export const clearHistory = async () => {
  try {
    await AsyncStorage.removeItem(KEYS.UPLOAD_HISTORY);
  } catch (e) {
    console.error('Failed to clear history', e);
  }
};
