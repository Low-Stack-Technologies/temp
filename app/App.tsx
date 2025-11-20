import React, { useState, useEffect } from 'react';
import {
  StyleSheet,
  Text,
  View,
  TextInput,
  TouchableOpacity,
  ScrollView,
  Alert,
  ActivityIndicator,
  Modal,
  Linking,
  StatusBar,
  Platform,
} from 'react-native';
import { LinearGradient } from 'expo-linear-gradient';
import { BlurView } from 'expo-blur';
import * as DocumentPicker from 'expo-document-picker';
import { Upload, File, X, CheckCircle, AlertCircle, Settings, Clock, Trash2, Copy, ExternalLink, ChevronDown } from 'lucide-react-native';
import { saveServerUrl, getServerUrl, saveHistoryItem, getHistory, clearHistory, HistoryItem, getServerConfig } from './storage';
import * as Clipboard from 'expo-clipboard';

interface FileItem {
  uri: string;
  name: string;
  size?: number;
  mimeType?: string;
  status: 'pending' | 'uploading' | 'success' | 'error';
  progress: number;
  downloadUrl?: string;
}

const ALL_TTL_OPTIONS = [
  { label: '5 Minutes', value: '300' },
  { label: '30 Minutes', value: '1800' },
  { label: '1 Hour', value: '3600' },
  { label: '6 Hours', value: '21600' },
  { label: '12 Hours', value: '43200' },
  { label: '1 Day', value: '86400' },
  { label: '1 Week', value: '604800' },
  { label: '1 Month', value: '2592000' },
];

// Design Constants
const COLORS = {
  bg: '#050505',
  primary: '#B026FF',
  secondary: '#00F0FF',
  text: '#ffffff',
  textMuted: '#8b9bb4',
  glassBorder: 'rgba(255, 255, 255, 0.1)',
  glassBg: 'rgba(255, 255, 255, 0.03)',
  success: '#00ff9d',
  error: '#ff0055',
};

export default function App() {
  const [serverUrl, setServerUrl] = useState('https://temp.low-stack.tech');
  const [ttl, setTtl] = useState('3600');
  const [files, setFiles] = useState<FileItem[]>([]);
  const [isUploading, setIsUploading] = useState(false);
  const [showSettings, setShowSettings] = useState(false);
  const [history, setHistory] = useState<HistoryItem[]>([]);
  const [showTTLSelector, setShowTTLSelector] = useState(false);
  
  const [minTtl, setMinTtl] = useState(60);
  const [maxTtl, setMaxTtl] = useState(2592000);

  useEffect(() => {
    loadSettings();
    loadHistory();
  }, []);

  useEffect(() => {
    if (serverUrl) {
      fetchConfig();
    }
  }, [serverUrl]);

  const loadSettings = async () => {
    const url = await getServerUrl();
    setServerUrl(url);
  };

  const fetchConfig = async () => {
    const config = await getServerConfig(serverUrl);
    if (config) {
      setMinTtl(config.min_ttl_seconds);
      setMaxTtl(config.max_ttl_seconds);
      
      const currentTtl = parseInt(ttl);
      if (currentTtl < config.min_ttl_seconds || currentTtl > config.max_ttl_seconds) {
        const validOption = ALL_TTL_OPTIONS.find(opt => {
          const val = parseInt(opt.value);
          return val >= config.min_ttl_seconds && val <= config.max_ttl_seconds;
        });
        if (validOption) {
          setTtl(validOption.value);
        } else {
          setTtl(config.min_ttl_seconds.toString());
        }
      }
    }
  };

  const loadHistory = async () => {
    const items = await getHistory();
    setHistory(items);
  };

  const handleServerUrlChange = (url: string) => {
    setServerUrl(url);
    saveServerUrl(url);
  };

  const handleClearHistory = async () => {
    Alert.alert(
      'Clear History',
      'Are you sure you want to clear your upload history?',
      [
        { text: 'Cancel', style: 'cancel' },
        { 
          text: 'Clear', 
          style: 'destructive', 
          onPress: async () => {
            await clearHistory();
            setHistory([]);
          } 
        }
      ]
    );
  };

  const pickFiles = async () => {
    try {
      const result = await DocumentPicker.getDocumentAsync({
        type: '*/*',
        multiple: true,
        copyToCacheDirectory: true,
      });

      if (!result.canceled) {
        const newFiles = result.assets.map((asset) => ({
          uri: asset.uri,
          name: asset.name,
          size: asset.size,
          mimeType: asset.mimeType,
          status: 'pending' as const,
          progress: 0,
        }));
        setFiles((prev) => [...prev, ...newFiles]);
      }
    } catch (err) {
      Alert.alert('Error', 'Failed to pick files');
    }
  };

  const removeFile = (index: number) => {
    setFiles((prev) => prev.filter((_, i) => i !== index));
  };

  const copyToClipboard = async (text: string) => {
    await Clipboard.setStringAsync(text);
    Alert.alert('Copied', 'Link copied to clipboard');
  };

  const uploadFiles = async () => {
    if (files.length === 0) {
      Alert.alert('Error', 'No files selected');
      return;
    }

    setIsUploading(true);

    const formData = new FormData();
    formData.append('ttl_seconds', ttl);

    files.forEach((file) => {
      // @ts-ignore
      formData.append('file', {
        uri: file.uri,
        name: file.name,
        type: file.mimeType || 'application/octet-stream',
      });
    });

    try {
      setFiles((prev) =>
        prev.map((f) => ({ ...f, status: 'uploading', progress: 0 }))
      );

      const xhr = new XMLHttpRequest();
      xhr.open('POST', `${serverUrl}/api/upload`);
      
      if (xhr.upload) {
        xhr.upload.onprogress = (event) => {
          if (event.lengthComputable) {
            const progress = event.loaded / event.total;
            setFiles((prev) =>
              prev.map((f) => ({ ...f, progress }))
            );
          }
        };
      }

      xhr.onload = async () => {
        if (xhr.status === 200) {
          const response = JSON.parse(xhr.responseText);
          if (response.success) {
             setFiles((prev) =>
              prev.map((f) => {
                const uploadedFile = response.files.find((uf: any) => uf.filename === f.name);
                return {
                  ...f,
                  status: 'success',
                  progress: 1,
                  downloadUrl: uploadedFile?.download_url,
                };
              })
            );
            
            for (const file of response.files) {
              const historyItem: HistoryItem = {
                id: file.id,
                filename: file.filename,
                size: file.size,
                downloadUrl: file.download_url,
                uploadedAt: Date.now(),
              };
              await saveHistoryItem(historyItem);
            }
            await loadHistory();
            
            Alert.alert('Success', 'Files uploaded successfully!');
          } else {
             setFiles((prev) => prev.map((f) => ({ ...f, status: 'error' })));
             Alert.alert('Upload Failed', response.error || 'Unknown error');
          }
        } else {
          setFiles((prev) => prev.map((f) => ({ ...f, status: 'error' })));
          Alert.alert('Error', `Server returned status ${xhr.status}`);
        }
        setIsUploading(false);
      };

      xhr.onerror = () => {
        setFiles((prev) => prev.map((f) => ({ ...f, status: 'error' })));
        Alert.alert('Error', 'Network request failed');
        setIsUploading(false);
      };

      xhr.send(formData);

    } catch (error) {
      setFiles((prev) => prev.map((f) => ({ ...f, status: 'error' })));
      Alert.alert('Error', 'An unexpected error occurred');
      setIsUploading(false);
    }
  };

  const getTTLLabel = () => {
    const option = ALL_TTL_OPTIONS.find(opt => opt.value === ttl);
    return option ? option.label : 'Custom';
  };

  const filteredTTLOptions = ALL_TTL_OPTIONS.filter(opt => {
    const val = parseInt(opt.value);
    return val >= minTtl && val <= maxTtl;
  });

  return (
    <View style={styles.container}>
      <StatusBar barStyle="light-content" />
      <LinearGradient
        colors={['#050505', '#0a0a0a', '#12051f']}
        style={StyleSheet.absoluteFill}
        start={{ x: 0, y: 0 }}
        end={{ x: 1, y: 1 }}
      />
      
      {/* Background Accents */}
      <View style={styles.bgAccent1} />
      <View style={styles.bgAccent2} />
      <BlurView intensity={80} tint="dark" style={StyleSheet.absoluteFill} />

      <View style={styles.headerContainer}>
        <View>
          <Text style={styles.title}>Temp Upload</Text>
          <Text style={styles.subtitle}>Secure file sharing</Text>
        </View>
        <TouchableOpacity onPress={() => setShowSettings(true)} style={styles.settingsButton}>
          <Settings color={COLORS.text} size={24} />
        </TouchableOpacity>
      </View>

      <ScrollView contentContainerStyle={styles.scrollContent} showsVerticalScrollIndicator={false}>
        <BlurView intensity={20} tint="dark" style={styles.card}>
          <Text style={styles.label}>FILE EXPIRATION</Text>
          <TouchableOpacity 
            style={styles.selectorButton}
            onPress={() => setShowTTLSelector(true)}
          >
            <Text style={styles.selectorText}>{getTTLLabel()}</Text>
            <ChevronDown color={COLORS.textMuted} size={20} />
          </TouchableOpacity>
        </BlurView>

        <TouchableOpacity
          style={styles.pickButton}
          onPress={pickFiles}
          disabled={isUploading}
        >
          <LinearGradient
            colors={[COLORS.primary, COLORS.secondary]}
            style={styles.gradientButton}
            start={{ x: 0, y: 0 }}
            end={{ x: 1, y: 1 }}
          >
            <File color="white" size={24} />
            <Text style={styles.buttonText}>Select Files</Text>
          </LinearGradient>
        </TouchableOpacity>

        {files.length > 0 && (
          <View style={styles.section}>
            <Text style={styles.sectionTitle}>Current Uploads</Text>
            {files.map((file, index) => (
              <BlurView intensity={30} tint="dark" key={index} style={styles.fileItem}>
                <View style={styles.fileIcon}>
                  <File color={COLORS.secondary} size={24} />
                </View>
                <View style={styles.fileInfo}>
                  <Text style={styles.fileName} numberOfLines={1}>
                    {file.name}
                  </Text>
                  <Text style={styles.fileSize}>
                    {file.size ? (file.size / 1024).toFixed(1) + ' KB' : 'Unknown size'}
                  </Text>
                  {file.status === 'uploading' && (
                    <View style={styles.progressBar}>
                      <LinearGradient
                        colors={[COLORS.primary, COLORS.secondary]}
                        start={{ x: 0, y: 0 }}
                        end={{ x: 1, y: 0 }}
                        style={[
                          styles.progressFill, 
                          { width: `${file.progress * 100}%` }
                        ]} 
                      />
                    </View>
                  )}
                </View>
                <View style={styles.fileActions}>
                  {file.status === 'success' ? (
                    <TouchableOpacity onPress={() => file.downloadUrl && copyToClipboard(file.downloadUrl)}>
                       <CheckCircle color={COLORS.success} size={20} />
                    </TouchableOpacity>
                  ) : file.status === 'error' ? (
                    <AlertCircle color={COLORS.error} size={20} />
                  ) : (
                    !isUploading && (
                      <TouchableOpacity onPress={() => removeFile(index)}>
                        <X color={COLORS.error} size={20} />
                      </TouchableOpacity>
                    )
                  )}
                </View>
              </BlurView>
            ))}
            
            <TouchableOpacity
              style={[styles.uploadButton, isUploading && styles.disabledButton]}
              onPress={uploadFiles}
              disabled={isUploading}
            >
              {isUploading ? (
                <ActivityIndicator color="white" />
              ) : (
                <LinearGradient
                  colors={[COLORS.success, '#00cc7d']}
                  style={styles.uploadGradient}
                  start={{ x: 0, y: 0 }}
                  end={{ x: 1, y: 1 }}
                >
                  <Upload color="white" size={24} style={{ marginRight: 8 }} />
                  <Text style={styles.uploadButtonText}>Upload {files.length} Files</Text>
                </LinearGradient>
              )}
            </TouchableOpacity>
          </View>
        )}

        {history.length > 0 && (
          <View style={styles.section}>
            <View style={styles.sectionHeader}>
              <Clock color={COLORS.textMuted} size={18} />
              <Text style={styles.sectionTitle}>History</Text>
            </View>
            {history.map((item) => (
              <BlurView intensity={20} tint="dark" key={item.id} style={styles.historyItem}>
                <View style={styles.fileInfo}>
                  <Text style={styles.fileName} numberOfLines={1}>{item.filename}</Text>
                  <Text style={styles.fileSize}>
                    {(item.size / 1024).toFixed(1)} KB • {new Date(item.uploadedAt).toLocaleDateString()}
                  </Text>
                </View>
                <View style={styles.historyActions}>
                  <TouchableOpacity onPress={() => copyToClipboard(item.downloadUrl)} style={styles.actionButton}>
                    <Copy color={COLORS.text} size={18} />
                  </TouchableOpacity>
                  <TouchableOpacity onPress={() => Linking.openURL(item.downloadUrl)} style={styles.actionButton}>
                    <ExternalLink color={COLORS.text} size={18} />
                  </TouchableOpacity>
                </View>
              </BlurView>
            ))}
          </View>
        )}
      </ScrollView>

      {/* Settings Modal */}
      <Modal
        visible={showSettings}
        animationType="slide"
        transparent={true}
        onRequestClose={() => setShowSettings(false)}
      >
        <BlurView intensity={90} tint="dark" style={styles.modalContainer}>
          <View style={styles.modalContent}>
            <View style={styles.modalHeader}>
              <Text style={styles.modalTitle}>Settings</Text>
              <TouchableOpacity onPress={() => setShowSettings(false)} style={styles.closeButton}>
                <X color={COLORS.text} size={24} />
              </TouchableOpacity>
            </View>

            <Text style={styles.label}>SERVER URL</Text>
            <TextInput
              style={styles.input}
              value={serverUrl}
              onChangeText={handleServerUrlChange}
              placeholder="https://temp.low-stack.tech"
              placeholderTextColor={COLORS.textMuted}
              autoCapitalize="none"
            />

            <TouchableOpacity style={styles.clearButton} onPress={handleClearHistory}>
              <Trash2 color={COLORS.error} size={20} style={{ marginRight: 8 }} />
              <Text style={styles.clearButtonText}>Clear History</Text>
            </TouchableOpacity>
          </View>
        </BlurView>
      </Modal>

      {/* TTL Selector Modal */}
      <Modal
        visible={showTTLSelector}
        animationType="fade"
        transparent={true}
        onRequestClose={() => setShowTTLSelector(false)}
      >
        <TouchableOpacity 
          style={styles.modalOverlay} 
          activeOpacity={1} 
          onPress={() => setShowTTLSelector(false)}
        >
          <BlurView intensity={50} tint="dark" style={styles.selectorModalContent}>
            <Text style={styles.selectorModalTitle}>Select Expiration</Text>
            {filteredTTLOptions.map((option) => (
              <TouchableOpacity
                key={option.value}
                style={[
                  styles.optionButton,
                  ttl === option.value && styles.selectedOption
                ]}
                onPress={() => {
                  setTtl(option.value);
                  setShowTTLSelector(false);
                }}
              >
                <Text style={[
                  styles.optionText,
                  ttl === option.value && styles.selectedOptionText
                ]}>
                  {option.label}
                </Text>
                {ttl === option.value && <CheckCircle color={COLORS.secondary} size={20} />}
              </TouchableOpacity>
            ))}
          </BlurView>
        </TouchableOpacity>
      </Modal>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: COLORS.bg,
  },
  bgAccent1: {
    position: 'absolute',
    top: -100,
    left: -100,
    width: 300,
    height: 300,
    borderRadius: 150,
    backgroundColor: COLORS.primary,
    opacity: 0.2,
  },
  bgAccent2: {
    position: 'absolute',
    bottom: -100,
    right: -100,
    width: 300,
    height: 300,
    borderRadius: 150,
    backgroundColor: COLORS.secondary,
    opacity: 0.2,
  },
  headerContainer: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    paddingTop: Platform.OS === 'ios' ? 60 : 40,
    paddingHorizontal: 24,
    paddingBottom: 24,
  },
  title: {
    fontSize: 32,
    fontWeight: '800',
    color: COLORS.text,
    letterSpacing: -1,
  },
  subtitle: {
    fontSize: 14,
    color: COLORS.textMuted,
    marginTop: 4,
    letterSpacing: 0.5,
  },
  settingsButton: {
    padding: 10,
    backgroundColor: 'rgba(255,255,255,0.05)',
    borderRadius: 14,
    borderWidth: 1,
    borderColor: COLORS.glassBorder,
  },
  scrollContent: {
    padding: 24,
    paddingTop: 0,
    paddingBottom: 40,
  },
  section: {
    marginTop: 32,
  },
  sectionHeader: {
    flexDirection: 'row',
    alignItems: 'center',
    marginBottom: 16,
    gap: 8,
  },
  sectionTitle: {
    fontSize: 18,
    fontWeight: '700',
    color: COLORS.text,
    letterSpacing: 0.5,
  },
  card: {
    borderRadius: 20,
    padding: 24,
    marginBottom: 24,
    borderWidth: 1,
    borderColor: COLORS.glassBorder,
    overflow: 'hidden',
  },
  label: {
    color: COLORS.textMuted,
    fontSize: 12,
    marginBottom: 12,
    fontWeight: '700',
    letterSpacing: 1,
  },
  input: {
    backgroundColor: 'rgba(0,0,0,0.3)',
    borderRadius: 12,
    padding: 16,
    color: COLORS.text,
    fontSize: 16,
    borderWidth: 1,
    borderColor: COLORS.glassBorder,
  },
  selectorButton: {
    backgroundColor: 'rgba(0,0,0,0.3)',
    borderRadius: 12,
    padding: 16,
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    borderWidth: 1,
    borderColor: COLORS.glassBorder,
  },
  selectorText: {
    color: COLORS.text,
    fontSize: 16,
    fontWeight: '500',
  },
  pickButton: {
    marginBottom: 12,
    borderRadius: 16,
    overflow: 'hidden',
    elevation: 8,
    shadowColor: COLORS.primary,
    shadowOffset: { width: 0, height: 4 },
    shadowOpacity: 0.3,
    shadowRadius: 12,
  },
  gradientButton: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    padding: 18,
  },
  buttonText: {
    color: 'white',
    fontSize: 18,
    fontWeight: '700',
    marginLeft: 10,
    letterSpacing: 0.5,
  },
  fileItem: {
    flexDirection: 'row',
    alignItems: 'center',
    padding: 16,
    borderRadius: 16,
    marginBottom: 12,
    overflow: 'hidden',
    borderWidth: 1,
    borderColor: COLORS.glassBorder,
  },
  fileIcon: {
    width: 48,
    height: 48,
    borderRadius: 12,
    backgroundColor: 'rgba(0, 240, 255, 0.1)',
    alignItems: 'center',
    justifyContent: 'center',
    marginRight: 16,
    borderWidth: 1,
    borderColor: 'rgba(0, 240, 255, 0.2)',
  },
  historyItem: {
    flexDirection: 'row',
    alignItems: 'center',
    padding: 16,
    borderRadius: 16,
    marginBottom: 12,
    overflow: 'hidden',
    borderWidth: 1,
    borderColor: COLORS.glassBorder,
  },
  fileInfo: {
    flex: 1,
    marginRight: 10,
  },
  fileName: {
    color: COLORS.text,
    fontSize: 16,
    fontWeight: '600',
    marginBottom: 4,
  },
  fileSize: {
    color: COLORS.textMuted,
    fontSize: 13,
  },
  fileActions: {
    padding: 4,
  },
  historyActions: {
    flexDirection: 'row',
    gap: 8,
  },
  actionButton: {
    padding: 8,
    backgroundColor: 'rgba(255,255,255,0.05)',
    borderRadius: 10,
    borderWidth: 1,
    borderColor: COLORS.glassBorder,
  },
  progressBar: {
    height: 4,
    backgroundColor: 'rgba(255,255,255,0.1)',
    borderRadius: 2,
    marginTop: 8,
    overflow: 'hidden',
  },
  progressFill: {
    height: '100%',
  },
  uploadButton: {
    borderRadius: 16,
    overflow: 'hidden',
    elevation: 8,
    marginTop: 12,
    shadowColor: COLORS.success,
    shadowOffset: { width: 0, height: 4 },
    shadowOpacity: 0.3,
    shadowRadius: 12,
  },
  uploadGradient: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    padding: 18,
  },
  disabledButton: {
    opacity: 0.7,
  },
  uploadButtonText: {
    color: 'white',
    fontSize: 18,
    fontWeight: '700',
  },
  modalContainer: {
    flex: 1,
    justifyContent: 'center',
    padding: 24,
  },
  modalContent: {
    backgroundColor: 'rgba(5,5,5,0.95)',
    borderRadius: 24,
    padding: 32,
    borderWidth: 1,
    borderColor: COLORS.glassBorder,
  },
  modalHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 32,
  },
  modalTitle: {
    fontSize: 28,
    fontWeight: '800',
    color: COLORS.text,
  },
  closeButton: {
    padding: 8,
    backgroundColor: 'rgba(255,255,255,0.05)',
    borderRadius: 12,
  },
  clearButton: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    padding: 16,
    borderRadius: 16,
    borderWidth: 1,
    borderColor: 'rgba(255, 0, 85, 0.3)',
    backgroundColor: 'rgba(255, 0, 85, 0.05)',
    marginTop: 32,
  },
  clearButtonText: {
    color: COLORS.error,
    fontSize: 16,
    fontWeight: '600',
  },
  modalOverlay: {
    flex: 1,
    backgroundColor: 'rgba(0,0,0,0.6)',
    justifyContent: 'center',
    padding: 24,
  },
  selectorModalContent: {
    backgroundColor: 'rgba(5,5,5,0.95)',
    borderRadius: 24,
    padding: 24,
    borderWidth: 1,
    borderColor: COLORS.glassBorder,
    overflow: 'hidden',
  },
  selectorModalTitle: {
    fontSize: 20,
    fontWeight: '700',
    color: COLORS.text,
    marginBottom: 20,
    textAlign: 'center',
    letterSpacing: 0.5,
  },
  optionButton: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    padding: 18,
    borderRadius: 16,
    marginBottom: 8,
    backgroundColor: 'rgba(255,255,255,0.03)',
    borderWidth: 1,
    borderColor: 'transparent',
  },
  selectedOption: {
    backgroundColor: 'rgba(0, 240, 255, 0.1)',
    borderColor: COLORS.secondary,
  },
  optionText: {
    fontSize: 16,
    color: COLORS.textMuted,
    fontWeight: '500',
  },
  selectedOptionText: {
    color: COLORS.text,
    fontWeight: '700',
  },
});
