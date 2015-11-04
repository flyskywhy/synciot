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
    public static void startSyncthing(Context ctx) {
        String syncthingPath = "/data/data/" + ctx.getApplicationContext().getPackageName() + "/syncthing";
        File file = new File(syncthingPath);
        if (!file.exists()) {
            try {
                AssetManager am = ctx.getApplicationContext().getAssets();
                InputStream is;
                is = am.open("syncthing");
                FileOutputStream out = new FileOutputStream(syncthingPath);
                byte[] buffer = new byte[1024 * 64];
                int read = is.read(buffer);

                while (read >= 0) {
                    out.write(buffer, 0, read);
                    read = is.read(buffer);
                }

                out.close();

                ShellInterface.runCommand("chmod 755 " + syncthingPath);

            } catch (IOException e) {
                e.printStackTrace();
            }
        }

        file = new File("/sdcard/iot/Sync");
        if (!file.exists() && !file.isDirectory()) {
            file.mkdir();
        }

        if (ShellInterface.isSuAvailable()) {
            Pattern RAMFS_PATTERN = Pattern.compile("SyncTemp ramfs");
            String out = ShellInterface.getProcessOutput("mount");
            Matcher matcher = RAMFS_PATTERN.matcher(out);
            if (!matcher.find()) {
                ShellInterface.runCommand("mkdir -p /sdcard/iot/SyncTemp/");
                ShellInterface.runCommand("mount -t ramfs -o mode=0777 none /sdcard/iot/SyncTemp/");
                ShellInterface.runCommand("touch /sdcard/iot/SyncTemp/.stfolder");
            }
        }

        final String config_xml = "/sdcard/iot/sync-config/config.xml";
        file = new File(config_xml);
        if (!file.exists()) {
            ShellInterface.runCommand("mkdir -p /sdcard/iot/sync-config/");
            ShellInterface.runCommand(syncthingPath + " -generate=/sdcard/iot/sync-config/");
        }

        String device_id = Unix4j.fromFile(config_xml)
                .grep("^        <device id=")
                .sed("s/^.*id=\"//")
                .sed("s/\">.*//")
                .toStringResult();

        String device_id_short = Unix4j.fromString(device_id)
                .sed("s/-.*//")
                .toStringResult();

        if ("0" != Unix4j.fromFile(config_xml).grep(GrepOption.count, "\"\\/data\\/Sync\"").toStringResult()) {
            Unix4j.fromFile(config_xml)
                    .sed("s/id=\"default\" path=\"\\/data\\/Sync\"/id=\"" + device_id_short + "\" path=\"\\/sdcard\\/iot\\/Sync\"/")
                    .sed("s/urAccepted>0/urAccepted>-1/")
                    .toFile(config_xml + ".tmp");
            ShellInterface.runCommand("mv " + config_xml + ".tmp " + config_xml);
        }
    }
}
