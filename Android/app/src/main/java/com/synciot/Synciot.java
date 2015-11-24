package com.synciot;

import android.content.Context;
import android.content.res.AssetManager;

import org.unix4j.Unix4j;
import org.unix4j.unix.grep.GrepOption;
import org.unix4j.unix.sed.SedOption;

import java.io.File;
import java.io.FileOutputStream;
import java.io.IOException;
import java.io.InputStream;
import java.util.regex.Matcher;
import java.util.regex.Pattern;

/**
 * Created by lizheng on 15-11-4.
 */
public class Synciot {

    private static Context CallerCtx;

    private static String dataPath;
    private static final String ORIGIN_SYNC_PATH = "/data/Sync";
    private static final String SYNC_PATH = "/sdcard/iot/Sync";
    private static final String SYNC_TEMP_PATH = "/sdcard/iot/SyncTemp";
    private static final String SYNCTHING_CONFIG_PATH = "/sdcard/iot/sync-config";

    private static final String ASSETS_SYNCTHING = "syncthing";
    private static String syncthing;
    private static final String ASSETS_SERVER_DEVICE_ID_TXT = "server_device_id.txt";
    private static final String SERVER_DEVICE_ID_TXT = SYNCTHING_CONFIG_PATH + "/" + ASSETS_SERVER_DEVICE_ID_TXT;
    private static final String CONFIG_XML = SYNCTHING_CONFIG_PATH + "/config.xml";

    private static String server_device_id;
    private static final String SERVER_DEFAULT_FOLDER_DEVICE = new StringBuilder()
            .append("        <device id=\"SERVER_DEVICE_ID\"></device>")
            .toString();
    private static final String SERVER_EXTRA_FOLDER_DEVICE = new StringBuilder()
            .append("    <folder id=\"FOLDER_ID\" path=\"FOLDER_PATH\" ro=\"false\" rescanIntervalS=\"60\" ignorePerms=\"false\" autoNormalize=\"false\">\n")
            .append("        <device id=\"CLIENT_DEVICE_ID\"></device>\n")
            .append("        <device id=\"SERVER_DEVICE_ID\"></device>\n")
            .append("        <minDiskFreePct>1</minDiskFreePct>\n")
            .append("        <versioning></versioning>\n")
            .append("        <copiers>0</copiers>\n")
            .append("        <pullers>0</pullers>\n")
            .append("        <hashers>0</hashers>\n")
            .append("        <order>random</order>\n")
            .append("        <ignoreDelete>false</ignoreDelete>\n")
            .append("    </folder>")
            .toString();

    private static final String SERVER_DEVICE = new StringBuilder()
            .append("    <device id=\"SERVER_DEVICE_ID\" name=\"Server\" compression=\"metadata\" introducer=\"false\">\n")
            .append("        <address>dynamic</address>\n")
            .append("    </device>")
            .toString();

    private static String device_id;
    private static String device_id_short;
    private static Thread syncthingThread;

    public static String getDevice_id() {
        return device_id;
    }

    public static String getDevice_id_short() {
        return device_id_short;
    }

    public static void startSyncthing(Context ctx) {
        // To get root at the very beginning
        ShellInterface.isSuAvailable();

        CallerCtx = ctx;
        dataPath = "/data/data/" + CallerCtx.getApplicationContext().getPackageName();
        syncthing = dataPath + "/" + ASSETS_SYNCTHING;
        File file = new File(syncthing);
        if (!file.exists()) {
            extractAssets(CallerCtx, ASSETS_SYNCTHING, syncthing);
            ShellInterface.runCommand("chmod 755 " + syncthing);
        }

        mkdirSync();
        mkdirSyncTemp();
        generateConfigXml();

        device_id = Unix4j.fromFile(CONFIG_XML)
                .grep("^        <device id=")
                .sed("s/^.*id=\"//")
                .sed("s/\">.*//")
                .toStringResult();

        device_id_short = Unix4j.fromString(device_id)
                .sed("s/-.*//")
                .toStringResult();

        file = new File(SERVER_DEVICE_ID_TXT);
        if (!file.exists()) {
            extractAssets(CallerCtx, ASSETS_SERVER_DEVICE_ID_TXT, SERVER_DEVICE_ID_TXT);
        }
        server_device_id = Unix4j.cat(SERVER_DEVICE_ID_TXT)
                .toStringResult();

        if (isOriginConfigXml()) {
            sedMisc2ConfigXml();
            sedSync2ConfigXml();
            sedSyncTemp2ConfigXml();
        }

        if (null == syncthingThread) {
            syncthingThread = new Thread(new Runnable() {
                public void run() {
                    ShellInterface.runCommand(syncthing
                            + " -no-browser -no-restart -gui-address=0.0.0.0:8384 -home="
                            + SYNCTHING_CONFIG_PATH + "/");
                }
            });
        }
        syncthingThread.start();
    }

    private static void sedSyncTemp2ConfigXml() {
        String server_extra_folder_device = Unix4j.fromString(SERVER_EXTRA_FOLDER_DEVICE)
                .sed(SedOption.substitute, "FOLDER_ID", device_id_short + "-Temp")
                .sed(SedOption.substitute, "FOLDER_PATH", SYNC_TEMP_PATH)
                .sed(SedOption.substitute, "CLIENT_DEVICE_ID", device_id)
                .sed(SedOption.substitute, "SERVER_DEVICE_ID", server_device_id)
                .toStringResult();
        Unix4j.fromFile(CONFIG_XML)
                .sed(SedOption.append, "^    </folder>", server_extra_folder_device)
                .toFile(CONFIG_XML + ".tmp");
        ShellInterface.runCommand("mv " + CONFIG_XML + ".tmp " + CONFIG_XML);
    }

    private static void sedSync2ConfigXml() {
        String server_default_folder_device = Unix4j.fromString(SERVER_DEFAULT_FOLDER_DEVICE)
                .sed(SedOption.substitute, "SERVER_DEVICE_ID", server_device_id)
                .toStringResult();
        Unix4j.fromFile(CONFIG_XML)
                .sed(SedOption.substitute, "id=\"default\" path=\"" + ORIGIN_SYNC_PATH + "\"", "id=\"" + device_id_short + "\" path=\"" + SYNC_PATH + "\"")
                .sed(SedOption.append, "^        <device id=", server_default_folder_device)
                .toFile(CONFIG_XML + ".tmp");
        ShellInterface.runCommand("mv " + CONFIG_XML + ".tmp " + CONFIG_XML);
    }

    private static void sedMisc2ConfigXml() {
        String server_device = Unix4j.fromString(SERVER_DEVICE)
                .sed(SedOption.substitute, "SERVER_DEVICE_ID", server_device_id)
                .toStringResult();
        Unix4j.fromFile(CONFIG_XML)
                .sed(SedOption.substitute, "localhost", device_id_short)
                .sed(SedOption.append, "^    </device>", server_device)
                .sed("s/urAccepted>0/urAccepted>-1/")
                .toFile(CONFIG_XML + ".tmp");
        ShellInterface.runCommand("mv " + CONFIG_XML + ".tmp " + CONFIG_XML);
    }

    private static boolean isOriginConfigXml() {
        final String count = Unix4j.fromFile(CONFIG_XML)
                .grep(GrepOption.count, "\"" + ORIGIN_SYNC_PATH + "\"")
                .toStringResult();
        return !("0".equals(count));
    }

    private static void generateConfigXml() {
        File file = new File(CONFIG_XML);
        if (!file.exists()) {
            ShellInterface.runCommand("mkdir -p " + SYNCTHING_CONFIG_PATH + "/");
            ShellInterface.runCommand(syncthing + " -generate=" + SYNCTHING_CONFIG_PATH + "/");
        }
    }

    private static void mkdirSyncTemp() {
        File file = new File(SYNC_TEMP_PATH);
        if (!file.exists() && !file.isDirectory()) {
            file.mkdir();
        }

        file =  new File(SYNC_TEMP_PATH + "/.stfolder");
        if (!file.exists()) {
            try {
                file.createNewFile();
            } catch (IOException e) {
                e.printStackTrace();
            }
        }

        if (ShellInterface.isSuAvailable()) {
            Pattern RAMFS_PATTERN = Pattern.compile("SyncTemp tmpfs");
            String out = ShellInterface.getProcessOutput("mount");
            Matcher matcher = RAMFS_PATTERN.matcher(out);
            if (!matcher.find()) {
                ShellInterface.runCommand("mount -t tmpfs -o mode=0777 none " + SYNC_TEMP_PATH + "/");
                ShellInterface.runCommand("touch " + SYNC_TEMP_PATH + "/.stfolder");
            }
        }
    }

    private static void mkdirSync() {
        File file = new File(SYNC_PATH);
        if (!file.exists() && !file.isDirectory()) {
            file.mkdir();
        }

        file =  new File(SYNC_PATH + "/.stfolder");
        if (!file.exists()) {
            try {
                file.createNewFile();
            } catch (IOException e) {
                e.printStackTrace();
            }
        }
    }

    private static void extractAssets(Context ctx, String assets, String path) {
        File file = new File(path);
        if (!file.exists()) {
            try {
                AssetManager am = ctx.getApplicationContext().getAssets();
                InputStream is;
                is = am.open(assets);
                FileOutputStream out = new FileOutputStream(path);
                byte[] buffer = new byte[1024 * 64];
                int read = is.read(buffer);

                while (read >= 0) {
                    out.write(buffer, 0, read);
                    read = is.read(buffer);
                }

                out.close();
            } catch (IOException e) {
                e.printStackTrace();
            }
        }
    }
}
