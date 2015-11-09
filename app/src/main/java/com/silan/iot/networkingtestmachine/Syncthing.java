package com.silan.iot.networkingtestmachine;

import android.content.Context;
import android.content.res.AssetManager;

import org.unix4j.Unix4j;
import org.unix4j.unix.grep.GrepOption;

import java.io.File;
import java.io.FileOutputStream;
import java.io.IOException;
import java.io.InputStream;
import java.util.regex.Matcher;
import java.util.regex.Pattern;

/**
 * Created by lizheng on 15-11-4.
 */
public class Syncthing {

    private static Context CallerCtx;

    private static String dataPath;
    private static final String SYNC_PATH = "/sdcard/iot/Sync";
    private static final String SYNC_PATH_SED = "\\/sdcard\\/iot\\/Sync";
    private static final String SYNC_TEMP_PATH = "/sdcard/iot/SyncTemp";
    private static final String SYNCTHING_CONFIG_PATH = "/sdcard/iot/sync-config";

    private static final String ASSETS_SYNCTHING = "syncthing";
    private static String syncthing;
    private static final String CONFIG_XML = SYNCTHING_CONFIG_PATH + "/config.xml";
    private static String device_id;
    private static String device_id_short;

    public static String getDevice_id() {
        return device_id;
    }

    public static String getDevice_id_short() {
        return device_id_short;
    }

    public static void startSyncthing(Context ctx) {
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

        if (isOriginConfigXml()) {
            sedMisc2ConfigXml();
            sedSync2ConfigXml();
        }
    }

    private static void sedSync2ConfigXml() {
        Unix4j.fromFile(CONFIG_XML)
                .sed("s/id=\"default\" path=\"\\/data\\/Sync\"/id=\"" + device_id_short + "\" path=\"" + SYNC_PATH_SED + "\"/")
                .toFile(CONFIG_XML + ".tmp");
        ShellInterface.runCommand("mv " + CONFIG_XML + ".tmp " + CONFIG_XML);
    }

    private static void sedMisc2ConfigXml() {
        Unix4j.fromFile(CONFIG_XML)
                .sed("s/urAccepted>0/urAccepted>-1/")
                .toFile(CONFIG_XML + ".tmp");
        ShellInterface.runCommand("mv " + CONFIG_XML + ".tmp " + CONFIG_XML);
    }

    private static boolean isOriginConfigXml() {
        final String count = Unix4j.fromFile(CONFIG_XML).grep(GrepOption.count, "\"/data/Sync\"").toStringResult();
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
        if (ShellInterface.isSuAvailable()) {
            Pattern RAMFS_PATTERN = Pattern.compile("SyncTemp ramfs");
            String out = ShellInterface.getProcessOutput("mount");
            Matcher matcher = RAMFS_PATTERN.matcher(out);
            if (!matcher.find()) {
                ShellInterface.runCommand("mkdir -p " + SYNC_TEMP_PATH + "/");
                ShellInterface.runCommand("mount -t ramfs -o mode=0777 none " + SYNC_TEMP_PATH + "/");
                ShellInterface.runCommand("touch " + SYNC_TEMP_PATH + "/.stfolder");
            }
        }
    }

    private static void mkdirSync() {
        File file = new File(SYNC_PATH);
        if (!file.exists() && !file.isDirectory()) {
            file.mkdir();
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
